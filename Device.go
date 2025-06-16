package onvif

import (
	"encoding/xml"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/beevik/etree"
	"github.com/ljhljh127/onvif/device"
	"github.com/ljhljh127/onvif/gosoap"
	"github.com/ljhljh127/onvif/networking"
	wsdiscovery "github.com/ljhljh127/onvif/ws-discovery"
)

// Xlmns XML Scheam
var Xlmns = map[string]string{
	"onvif":   "http://www.onvif.org/ver10/schema",
	"tds":     "http://www.onvif.org/ver10/device/wsdl",
	"trt":     "http://www.onvif.org/ver10/media/wsdl",
	"tev":     "http://www.onvif.org/ver10/events/wsdl",
	"tptz":    "http://www.onvif.org/ver20/ptz/wsdl",
	"timg":    "http://www.onvif.org/ver20/imaging/wsdl",
	"tan":     "http://www.onvif.org/ver20/analytics/wsdl",
	"xmime":   "http://www.w3.org/2005/05/xmlmime",
	"wsnt":    "http://docs.oasis-open.org/wsn/b-2",
	"xop":     "http://www.w3.org/2004/08/xop/include",
	"wsa":     "http://www.w3.org/2005/08/addressing",
	"wstop":   "http://docs.oasis-open.org/wsn/t-1",
	"wsntw":   "http://docs.oasis-open.org/wsn/bw-2",
	"wsrf-rw": "http://docs.oasis-open.org/wsrf/rw-2",
	"wsaw":    "http://www.w3.org/2006/05/addressing/wsdl",
}

// DeviceType alias for int
type DeviceType int

// Onvif Device Tyoe
const (
	NVD DeviceType = iota
	NVS
	NVA
	NVT
)

func (devType DeviceType) String() string {
	stringRepresentation := []string{
		"NetworkVideoDisplay",
		"NetworkVideoStorage",
		"NetworkVideoAnalytics",
		"NetworkVideoTransmitter",
	}
	i := uint8(devType)
	switch {
	case i <= uint8(NVT):
		return stringRepresentation[i]
	default:
		return strconv.Itoa(int(i))
	}
}

// DeviceInfo struct contains general information about ONVIF device
type DeviceInfo struct {
	Manufacturer    string
	Model           string
	FirmwareVersion string
	SerialNumber    string
	HardwareId      string
}

// Device for a new device of onvif and DeviceInfo
// struct represents an abstract ONVIF device.
// It contains methods, which helps to communicate with ONVIF device
type Device struct {
	params    DeviceParams
	endpoints map[string]string
	info      DeviceInfo
}

type DeviceParams struct {
	Xaddr      string
	Username   string
	Password   string
	HttpClient *http.Client
}

// GetServices return available endpoints
func (dev *Device) GetServices() map[string]string {
	return dev.endpoints
}

// GetDeviceInfo return available endpoints
func (dev *Device) GetDeviceInfo() DeviceInfo {
	return dev.info
}

// GetDeviceParams return available endpoints
func (dev *Device) GetDeviceParams() DeviceParams {
	return dev.params
}

func readResponse(resp *http.Response) string {
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	return string(b)
}

// GetAvailableDevicesAtSpecificEthernetInterface ...
func GetAvailableDevicesAtSpecificEthernetInterface(interfaceName string) ([]Device, error) {
	// Call a ws-discovery Probe Message to Discover NVT type Devices
	devices, err := wsdiscovery.SendProbe(interfaceName, nil, []string{"dn:" + NVT.String()}, map[string]string{"dn": "http://www.onvif.org/ver10/network/wsdl"})
	if err != nil {
		return nil, err
	}

	nvtDevicesSeen := make(map[string]bool)
	nvtDevices := make([]Device, 0)

	for _, j := range devices {
		doc := etree.NewDocument()
		if err := doc.ReadFromString(j); err != nil {
			return nil, err
		}

		//  ProbeMatch 전체를 순회하며 탐색
		for _, probeMatch := range doc.Root().FindElements("./Body/ProbeMatches/ProbeMatch") {
			// XAddr 추출
			xaddrElement := probeMatch.FindElement("./XAddrs")
			if xaddrElement == nil {
				continue
			}

			xaddrText := xaddrElement.Text()
			xaddr := strings.Split(strings.Split(xaddrText, " ")[0], "/")[2]
			if nvtDevicesSeen[xaddr] {
				continue
			}

			//  Device 생성
			dev, err := NewDevice(DeviceParams{Xaddr: strings.Split(xaddr, " ")[0]})
			if err != nil {
				// TODO(jfsmig) print a warning
				continue
			}

			//  Scopes에서 디바이스 정보 파싱해서 dev.info에 저장
			if scopesElement := probeMatch.FindElement("./Scopes"); scopesElement != nil {
				scopeText := scopesElement.Text()
				parseDeviceInfoFromScopes(dev, scopeText)
			}

			nvtDevicesSeen[xaddr] = true
			nvtDevices = append(nvtDevices, *dev)
		}
	}

	return nvtDevices, nil
}

func parseDeviceInfoFromScopes(dev *Device, scopeText string) {
	// 제조사명 추출
	nameRegex := regexp.MustCompile(`onvif://www\.onvif\.org/name/([A-Za-z0-9_-]+)`)
	if match := nameRegex.FindStringSubmatch(scopeText); len(match) > 1 {
		dev.info.Manufacturer = match[1]
	}

	// 모델명 추출
	hardwareRegex := regexp.MustCompile(`onvif://www\.onvif\.org/hardware/([A-Za-z0-9_-]+)`)
	if match := hardwareRegex.FindStringSubmatch(scopeText); len(match) > 1 {
		dev.info.Model = match[1]
	}

	// 하드웨어 ID 추출
	hwIdRegex := regexp.MustCompile(`onvif://www\.onvif\.org/hardware/([A-Za-z0-9_-]+)`)
	if match := hwIdRegex.FindStringSubmatch(scopeText); len(match) > 1 {
		dev.info.HardwareId = match[1]
	}

	// MAC 주소 추출
	macRegex := regexp.MustCompile(`onvif://www\.onvif\.org/MAC/([A-Fa-f0-9:_-]+)`)
	if match := macRegex.FindStringSubmatch(scopeText); len(match) > 1 {
		// MAC을 SerialNumber에 저장하거나 별도 필드 추가
		dev.info.SerialNumber = match[1]
	}

	// 위치 정보나 기타 정보를 FirmwareVersion에 임시 저장할 수도 있음
	locationRegex := regexp.MustCompile(`onvif://www\.onvif\.org/location/([A-Za-z0-9_-]+)`)
	if match := locationRegex.FindStringSubmatch(scopeText); len(match) > 1 {
		dev.info.FirmwareVersion = match[1]
	}
}

func (dev *Device) getSupportedServices(resp *http.Response) error {
	doc := etree.NewDocument()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	resp.Body.Close()

	if err := doc.ReadFromBytes(data); err != nil {
		//log.Println(err.Error())
		return err
	}

	services := doc.FindElements("./Envelope/Body/GetCapabilitiesResponse/Capabilities/*/XAddr")
	for _, j := range services {
		dev.addEndpoint(j.Parent().Tag, j.Text())
	}

	extension_services := doc.FindElements("./Envelope/Body/GetCapabilitiesResponse/Capabilities/Extension/*/XAddr")
	for _, j := range extension_services {
		dev.addEndpoint(j.Parent().Tag, j.Text())
	}

	return nil
}

// NewDevice function construct a ONVIF Device entity
func NewDevice(params DeviceParams) (*Device, error) {
	dev := new(Device)
	dev.params = params
	dev.endpoints = make(map[string]string)
	dev.addEndpoint("Device", "http://"+dev.params.Xaddr+"/onvif/device_service")

	if dev.params.HttpClient == nil {
		dev.params.HttpClient = new(http.Client)
	}

	getCapabilities := device.GetCapabilities{Category: "All"}

	resp, err := dev.CallMethod(getCapabilities)

	if err != nil || resp.StatusCode != http.StatusOK {
		return nil, errors.New("camera is not available at " + dev.params.Xaddr + " or it does not support ONVIF services")
	}

	err = dev.getSupportedServices(resp)
	if err != nil {
		return nil, err
	}

	return dev, nil
}

func (dev *Device) addEndpoint(Key, Value string) {
	//use lowCaseKey
	//make key having ability to handle Mixed Case for Different vendor devcie (e.g. Events EVENTS, events)
	lowCaseKey := strings.ToLower(Key)

	// Replace host with host from device params.
	if u, err := url.Parse(Value); err == nil {
		u.Host = dev.params.Xaddr
		Value = u.String()
	}

	dev.endpoints[lowCaseKey] = Value
}

// GetEndpoint returns specific ONVIF service endpoint address
func (dev *Device) GetEndpoint(name string) string {
	return dev.endpoints[name]
}

func (dev Device) buildMethodSOAP(msg string) (gosoap.SoapMessage, error) {
	doc := etree.NewDocument()
	if err := doc.ReadFromString(msg); err != nil {
		//log.Println("Got error")

		return "", err
	}
	element := doc.Root()

	soap := gosoap.NewEmptySOAP()
	soap.AddBodyContent(element)

	return soap, nil
}

// getEndpoint functions get the target service endpoint in a better way
func (dev Device) getEndpoint(endpoint string) (string, error) {

	// common condition, endpointMark in map we use this.
	if endpointURL, bFound := dev.endpoints[endpoint]; bFound {
		return endpointURL, nil
	}

	//but ,if we have endpoint like event、analytic
	//and sametime the Targetkey like : events、analytics
	//we use fuzzy way to find the best match url
	var endpointURL string
	for targetKey := range dev.endpoints {
		if strings.Contains(targetKey, endpoint) {
			endpointURL = dev.endpoints[targetKey]
			return endpointURL, nil
		}
	}
	return endpointURL, errors.New("target endpoint service not found")
}

// CallMethod functions call an method, defined <method> struct.
// You should use Authenticate method to call authorized requests.
func (dev Device) CallMethod(method interface{}) (*http.Response, error) {
	pkgPath := strings.Split(reflect.TypeOf(method).PkgPath(), "/")
	pkg := strings.ToLower(pkgPath[len(pkgPath)-1])

	endpoint, err := dev.getEndpoint(pkg)
	if err != nil {
		return nil, err
	}
	return dev.callMethodDo(endpoint, method)
}

// CallMethod functions call an method, defined <method> struct with authentication data
func (dev Device) callMethodDo(endpoint string, method interface{}) (*http.Response, error) {
	output, err := xml.MarshalIndent(method, "  ", "    ")
	if err != nil {
		return nil, err
	}

	soap, err := dev.buildMethodSOAP(string(output))
	if err != nil {
		return nil, err
	}

	soap.AddRootNamespaces(Xlmns)
	soap.AddAction()

	//Auth Handling
	if dev.params.Username != "" && dev.params.Password != "" {
		soap.AddWSSecurity(dev.params.Username, dev.params.Password)
	}

	return networking.SendSoap(dev.params.HttpClient, endpoint, soap.String())
}
