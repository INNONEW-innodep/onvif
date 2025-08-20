package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/ljhljh127/onvif"
	"github.com/ljhljh127/onvif/device"
	"github.com/ljhljh127/onvif/media"
	onvifXSD "github.com/ljhljh127/onvif/xsd/onvif"
)

type Service struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type DeviceServices struct {
	Xaddr    string    `json:"xaddr"`
	Services []Service `json:"services"`
}

type DeviceInfoResponse struct {
	Manufacturer    string `xml:"Manufacturer"`
	Model           string `xml:"Model"`
	FirmwareVersion string `xml:"FirmwareVersion"`
	SerialNumber    string `xml:"SerialNumber"`
	HardwareId      string `xml:"HardwareId"`
}

// 서비스 구조체화
func parseServices(servicesMap map[string]string, xaddr string) DeviceServices {
	var deviceServices DeviceServices
	deviceServices.Xaddr = xaddr

	for name, url := range servicesMap {
		deviceServices.Services = append(deviceServices.Services, Service{
			Name: name,
			URL:  url,
		})
	}
	return deviceServices
}

// 활성화된 네트워크중 도커 네트워크 제외
func getPhysicalActiveInterfaces() []net.Interface {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil
	}

	var physicalIfaces []net.Interface

	for _, iface := range ifaces {

		isUp := iface.Flags&net.FlagUp != 0
		isRunning := iface.Flags&net.FlagRunning != 0
		isLoopback := iface.Flags&net.FlagLoopback != 0

		isDocker := strings.HasPrefix(iface.Name, "docker") ||
			strings.HasPrefix(iface.Name, "br-") ||
			strings.HasPrefix(iface.Name, "veth")

		if isUp && isRunning && !isLoopback && !isDocker {
			physicalIfaces = append(physicalIfaces, iface)
		}
	}

	return physicalIfaces
}

// 활성화된 모든 네트워크 네임 리스트 반환
func getPhysicalActiveInterfaceNames() []string {
	physicalIfaces := getPhysicalActiveInterfaces()
	names := make([]string, len(physicalIfaces))

	for i, iface := range physicalIfaces {
		names[i] = iface.Name
	}

	return names
}

// 연결된 네트워크 인터페이스들에서 onvif장비를 찾아 가능한서비스 목록과 함께 반환
func getDeviceService() []DeviceServices {
	var deviceList []DeviceServices
	networkInterfaces := getPhysicalActiveInterfaceNames()
	fmt.Printf("Connected Interface List %v", networkInterfaces)
	for _, networkinterface := range networkInterfaces {

		devices, err := onvif.GetAvailableDevicesAtSpecificEthernetInterface(networkinterface)
		if err != nil {
			fmt.Printf("Get Device From %v error!", networkinterface)
			return nil
		}

		for _, dev := range devices {
			deviceEndpoint := dev.GetEndpoint("device")
			if err != nil {
				fmt.Println("Error getting device endpoint:", err)
				continue
			}

			u, err := url.Parse(deviceEndpoint)
			if err != nil {
				fmt.Println("Error parsing device endpoint:", err)
				continue
			}
			xaddr := u.Host

			servicesMap := dev.GetServices()
			deviceServices := parseServices(servicesMap, xaddr)

			// 디바이스 정보 추가
			deviceInfo := dev.GetDeviceInfo()
			fmt.Printf("Manufacturer: %s, Model: %s\n",
				deviceInfo.Manufacturer, deviceInfo.Model)

			deviceList = append(deviceList, deviceServices)

		}
	}
	return deviceList
}

// 주어진 로그인 정보를 포함하여 onvif 디바이스 객체를 생성
func newOnvifDevice(xaddr, username, passwd string) (*onvif.Device, error) {
	device, err := onvif.NewDevice(onvif.DeviceParams{
		Xaddr:      xaddr,
		Username:   username,
		Password:   passwd,
		TimeOffset: 0,
	})
	if err != nil {
		return nil, err
	}
	return device, nil
}

// 카메라 정보 가져오는 함수
func getDeviceInfo(onvifDevice *onvif.Device) (*DeviceInfoResponse, error) {
	fmt.Println("\nFetching device information...")
	resp, err := onvifDevice.CallMethod(device.GetDeviceInformation{})
	if err != nil {
		fmt.Println("Error calling GetDeviceInformation:", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error: unexpected status code %v", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return nil, err
	}
	if len(body) == 0 {
		fmt.Println("Error: Empty response body")
		return nil, err
	}

	// Device 정보 파싱
	var devInfoResp struct {
		DeviceInfo DeviceInfoResponse `xml:"Body>GetDeviceInformationResponse"`
	}
	if err := xml.Unmarshal(body, &devInfoResp); err != nil {
		fmt.Println("Error unmarshaling device info:", err)
		fmt.Println("Response body:", string(body))
		return nil, err
	}

	return &devInfoResp.DeviceInfo, nil
}

// 카메라 프로파일을 가져오는 함수
func GetDeviceProfiles(onvifDevice *onvif.Device) ([]onvifXSD.Profile, error) {
	fmt.Println("\nFetching profiles...")
	resp, err := onvifDevice.CallMethod(media.GetProfiles{})
	if err != nil {
		fmt.Println("Error calling GetProfiles:", err)
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		fmt.Println("Error: GetProfiles unexpected status code:", resp.StatusCode)
		body, _ := io.ReadAll(resp.Body)
		fmt.Println("Response body:", string(body))
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading GetProfiles response body:", err)
		return nil, err
	}
	if len(body) == 0 {
		fmt.Println("Error: Empty GetProfiles response body")
		return nil, err
	}

	var profilesResp struct {
		Profiles []onvifXSD.Profile `xml:"Body>GetProfilesResponse>Profiles"`
	}

	if err := xml.Unmarshal(body, &profilesResp); err != nil {
		fmt.Println("Error unmarshaling profiles:", err)
		fmt.Println("Response body:", string(body))
		return nil, err
	}

	fmt.Println("Found", len(profilesResp.Profiles), "profiles:")

	return profilesResp.Profiles, nil
}

// 해당 프로파일의 RTSP 주소를 가져오는 함수
func getRTSPUrlFromProfile(onvifDevice *onvif.Device, profile onvifXSD.Profile) (string, error) {
	streamReq := media.GetStreamUri{
		StreamSetup: onvifXSD.StreamSetup{
			Stream: "RTP-Unicast",
			Transport: onvifXSD.Transport{
				Protocol: "RTSP",
			},
		},
		ProfileToken: profile.Token,
	}

	resp, err := onvifDevice.CallMethod(streamReq)
	if err != nil {
		fmt.Println("Error calling GetStreamUri for profile", profile.Token, ":", err)
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("Error: GetStreamUri unexpected status code:", resp.StatusCode)
		body, _ := io.ReadAll(resp.Body)
		fmt.Println("Response body:", string(body))
		return "", err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading GetStreamUri response body:", err)
		return "", err
	}
	if len(body) == 0 {
		fmt.Println("Error: Empty GetStreamUri response body")
		return "", err
	}

	var streamUriResp struct {
		Uri string `xml:"Body>GetStreamUriResponse>MediaUri>Uri"`
	}
	if err := xml.Unmarshal(body, &streamUriResp); err != nil {
		fmt.Println("Error unmarshaling stream URI:", err)
		fmt.Println("Response body:", string(body))
		return "", err
	}
	return streamUriResp.Uri, nil
}

func main() {
	// 온비프 디바이스 서비스 호출 부
	deviceList := getDeviceService()
	if len(deviceList) == 0 {
		fmt.Println("No devices to summarize.")
		return
	}

	// 디바이스 정보 출력 부
	fmt.Println("\nSummary of Devices:")
	for i, dev := range deviceList {
		fmt.Printf("[%d] IP:%s Service Count:%d \n", i, dev.Xaddr, len(dev.Services))
	}

	// 디바이스 선택 부 추 후 제거
	fmt.Print("\nSelect device number (0 to ", len(deviceList)-1, "): ")
	var deviceIndex int
	_, err := fmt.Scan(&deviceIndex)
	if err != nil || deviceIndex < 0 || deviceIndex >= len(deviceList) {
		fmt.Println("Invalid device number.")
		return
	}
	selectedXaddr := deviceList[deviceIndex].Xaddr
	fmt.Println("Selected device:", selectedXaddr)

	fmt.Print("Enter username (leave blank if none): ")
	var username string
	fmt.Scanln(&username)
	username = strings.TrimSpace(username)

	fmt.Print("Enter password (leave blank if none): ")
	var password string
	fmt.Scanln(&password)
	password = strings.TrimSpace(password)

	// 디바이스 생성 부
	onvifDevice, err := newOnvifDevice(selectedXaddr, username, password)
	if err != nil {
		fmt.Println("Error creating device:", err)
		return
	}

	// 카메라 시간 동기화
	onvifDevice.SyncTimeWithCamera()

	// 카메라 정보 가져오는 부
	deviceInfo, err := getDeviceInfo(onvifDevice)
	if err != nil {
		fmt.Printf("Error occured in get camera info %v\n", err)
		return
	}

	// 카메라 정보 출력부
	fmt.Println("Manufacturer:", deviceInfo.Manufacturer)
	fmt.Println("Model:", deviceInfo.Model)
	fmt.Println("FirmwareVersion:", deviceInfo.FirmwareVersion)
	fmt.Println("SerialNumber:", deviceInfo.SerialNumber)
	fmt.Println("HardwareId:", deviceInfo.HardwareId)

	// 카메라 프로파일 가져오는 부
	profiles, err := GetDeviceProfiles(onvifDevice)
	if err != nil {
		fmt.Println("Error occured in get camera profiles")
		return
	}

	for _, profile := range profiles {
		rtspUrl, err := getRTSPUrlFromProfile(onvifDevice, profile)
		if err != nil {
			fmt.Println("Error occured in get rtsp url")
		}

		fmt.Println("-----------------------------------------------------------------------")
		fmt.Printf("Device %v Profile %v Info\n", selectedXaddr, profile.Name)
		fmt.Printf("Encoding %v\n", profile.VideoEncoderConfiguration.Encoding)
		fmt.Printf("Resolution Width %v\n", profile.VideoEncoderConfiguration.Resolution.Width)
		fmt.Printf("Resolution Height %v\n", profile.VideoEncoderConfiguration.Resolution.Height)
		fmt.Printf("FPS Limit %v\n", profile.VideoEncoderConfiguration.RateControl.FrameRateLimit)
		fmt.Printf("Bitrate Limit %v\n", profile.VideoEncoderConfiguration.RateControl.BitrateLimit)
		fmt.Printf("RTSP URL %v\n", rtspUrl)
	}
}

// TO DO: 추후 IP 대역에 대한 처리를 추가로 진행
// // ___________________________________________________________________________
// package main

// import (
// 	"bytes"
// 	"fmt"
// 	"log"
// 	"net"
// 	"sync"

// 	"github.com/ljhljh127/onvif"
// )

// func ScanOnvifDevicesInRange(startIP, endIP string) ([]onvif.Device, error) {
// 	var devices []onvif.Device
// 	var mu sync.Mutex
// 	var wg sync.WaitGroup

// 	ipStart := net.ParseIP(startIP).To4()
// 	ipEnd := net.ParseIP(endIP).To4()
// 	if ipStart == nil || ipEnd == nil {
// 		return nil, fmt.Errorf("invalid IP range")
// 	}

// 	for ip := ipStart; bytes.Compare(ip, ipEnd) <= 0; incIP(ip) {
// 		wg.Add(1)
// 		ipCopy := make(net.IP, len(ip))
// 		copy(ipCopy, ip)
// 		go func(ipStr string) {
// 			defer wg.Done()
// 			dev, err := onvif.NewDevice(onvif.DeviceParams{Xaddr: ipStr})
// 			if err != nil {
// 				return
// 			}
// 			mu.Lock()
// 			devices = append(devices, *dev)
// 			mu.Unlock()
// 		}(ipCopy.String())
// 	}

// 	wg.Wait()
// 	return devices, nil
// }

// func incIP(ip net.IP) {
// 	for j := len(ip) - 1; j >= 0; j-- {
// 		ip[j]++
// 		if ip[j] > 0 {
// 			break
// 		}
// 	}
// }

// func main() {
// 	devices, err := ScanOnvifDevicesInRange("172.20.10.0", "172.20.10.254")
// 	if err != nil {
// 		log.Fatalf("Error: %v", err)
// 	}
// 	if len(devices) == 0 {
// 		fmt.Println("No devices found")
// 	}
// 	for _, dev := range devices {
// 		fmt.Printf("Found device: %s\n", dev)
// 	}
// }
