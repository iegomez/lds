## Loraserver device simulator

This is an utility program to simulate ABP devices uplinks for the [loraserver](https://loraserver.io) project.  Basically, it acts as if a `lora-gateway-bridge` had received a packet and publishes through MQTT for a corresponding `loraserver` to receive it.
It supports all bands and configurations LoRaWAN versions 1.0 and 1.1.  

It has a simple but complete GUI (built with https://github.com/andlabs/ui) that allows to configure everything that's needed, such as MQTT broker and credentials, device keys, LoRaWAN version, message marshaling method, data payload, etc.  

The GUI allows to modify all options, but they may also be seeded with a conf file for ease of use. An example file is provided:

```toml
#Configuration.
[mqtt]
server="tcp://localhost:1883"
user="your-user"
password="your-password"

[gateway]
mac="your-gw-mac"

[band]
name="US_902_928"

[device]
eui="0000000000000001"
address="00815bd9"
network_session_encription_key="aee1da4b88979ae4f75475ff0db51c04"
serving_network_session_integrity_key="e0422b2a0307f24b2986e5b24ca8d3d9"
forwarding_network_session_integrity_key="bc97ea1ff62e7a3490135d989aae6bca"
application_session_key="037882c03dd6b20724b44d623abb4f95"
marshaler="json"
nwk_key="00000000000000010000000000000001"
app_key="00000000000000010000000000000001"
major=0
mac_version=1

[data_rate]
bandwith=125
spread_factor=10
bit_rate=0

[rx_info]
channel=0
code_rate="4/5"
crc_status=1
frequency=902300000
lora_snr=7.0
rf_chain=1
rssi=-57

[raw_payload]
payload="ff00"
use_payload=true

[default_data]
names = ["Temp", "Lat", "Lng"]
data = [[25.0, 127.0, 2.0], [-33.4348474, 90.0, 4.0], [-70.6157308, 180.0, 4.0]]
```

### Data

The data to be sent may be presented as a hex string representation of the raw bytes, or using our encoding method (which then needs to be decoded accordingly at `lora-app-server`). As a reference, this is how we encode our data:

```go
func GenerateFloat(originalFloat, maxValue float32, numBytes int32) []byte {
	byteArray := make([]byte, numBytes)
	if numBytes == 4 {
		encodedFloat := uint32((originalFloat / maxValue) * float32(math.Pow(2, 31)))
		binary.BigEndian.PutUint32(byteArray, encodedFloat)
	} else if numBytes == 3 {
		encodedFloat := uint32((originalFloat / maxValue) * float32(math.Pow(2, 23)))
		temp := make([]byte, 4)
		binary.BigEndian.PutUint32(temp, encodedFloat)
		byteArray = temp[1:]
	} else if numBytes == 2 {
		encodedFloat := uint16((originalFloat / maxValue) * float32(math.Pow(2, 15)))
		binary.BigEndian.PutUint16(byteArray, encodedFloat)
	} else if numBytes == 1 {
		byteArray[0] = uint8(originalFloat)
	}
	return byteArray
}

func GenerateInt(originalInt, numBytes int32) []byte {

	bRep := make([]byte, numBytes)
	if numBytes == 4 {
		binary.BigEndian.PutUint32(bRep, uint32(originalInt))
	} else if numBytes == 3 {
		temp := make([]byte, 4)
		binary.BigEndian.PutUint32(temp, uint32(originalInt))
		bRep = temp[1:]
	} else if numBytes == 2 {
		binary.BigEndian.PutUint16(bRep, uint16(originalInt))
	} else if numBytes == 1 {
		bRep[0] = uint8(originalInt)
	}

	return bRep
}
```

Values may be added using the `Add value` button and setting the options:

![](images/data_encoding.png?raw=true)

### Building

The GUI is built using https://github.com/andlabs/ui, so please check that repo to see requitrements for your system. Once those are met, you may build the package like this: 

```
make dev-requirements
make requirements
make
```