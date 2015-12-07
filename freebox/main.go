package main
import (
"fmt"
"flag"
"./mdns"
"errors"
"log"
"strings"
"net/http"
	"io/ioutil"
)

// Command-line flags.
var (
	freeboxIpAddr   = flag.String("ip", "", "Freebox ip")
	apiBaseUrl   = flag.String("apiBaseUrl", "", "Freebox api base url")
	autoDetect   = flag.Bool("autoDetect", false, "autoDetect Freebox ip?")
)

func autoDetectIp() (ip *string, apiBaseUrl *string, err2 error){

	// Make a channel for results and start listening
	entriesCh := make(chan *mdns.ServiceEntry, 8)
	// Start the lookup
	err := mdns.Lookup("_fbx-api._tcp", entriesCh)
	if err != nil {
		fmt.Println(err)
		return nil, nil, err
	}

	for entry := range entriesCh {
		var ipv4 = entry.AddrV4.String()
		fmt.Printf("Info = %s\n",entry.Info)
		var majorVersion = ""
		var apiBaseUrl = ""
		for i := 0; i < len(entry.InfoFields); i++ {
			keyValue := strings.Split(entry.InfoFields[i],"=")
			if keyValue[0] == "api_version" {
				majorVersion = strings.Split(keyValue[1],".")[0]
			} else if keyValue[0] == "api_base_url" {
				apiBaseUrl = keyValue[1]
			}

			if  majorVersion != "" && apiBaseUrl != "" {
				var url = apiBaseUrl + "v" + majorVersion +"/"
				return &ipv4, &url , nil
			}
		}
		return nil,nil, errors.New("apiBaseUrl or majorVersion not found")
	}

	close(entriesCh)
	return nil, nil, errors.New("Freebox not found")
}

func main() {
	flag.Parse()
	if  (*freeboxIpAddr == "" || *apiBaseUrl == "") && *autoDetect == false {
		flag.Usage()
		return
	}
	if *freeboxIpAddr == "" || *apiBaseUrl == ""{
		tmp, tmp2, err  := autoDetectIp()
		if err != nil {
			log.Printf("[ERR] %v\n",err);
			return
		}
		freeboxIpAddr = tmp
		apiBaseUrl = tmp2
	}
	fmt.Printf("Ip:%s\n",*freeboxIpAddr)
	fmt.Printf("ApiBaseUrl:%s\n",*apiBaseUrl)


	url := "http://"+*freeboxIpAddr+*apiBaseUrl+"login/";
	fmt.Println("URL:>", url)

	req, err := http.NewRequest("GET", url,nil)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))
}

