## Loraserver device simulator

This is an utility program to simulate devices for the [loraserver](https://loraserver.io) project.  Basically, it acts as a `lora-gateway-bridge` middleman, publishing and receiving messages through MQTT.
It supports all bands and configurations LoRaWAN versions 1.0 and 1.1.  

It has a simple but complete GUI built with [imgui-go](https://github.com/inkyblackness/imgui-go) and OpenGL 3.2, that allows to configure everything that's needed, such as MQTT broker and credentials, device keys, LoRaWAN version, message marshaling method, data payload, etc.

**Important**: This is a work in progress. LoRaWAN 1.1 downlinks are broken right now (need to fix MIC validation), there may be other bugs too, the `cli` mode needs to be rewritten, etc.


### Conf

The GUI allows to modify all options, but they may also be seeded with a conf file for ease of use. An example file is provided to get an idea, but the program will only load a conf file with the name `conf.toml` located at the same dir as the binary, or if the path is given with the `--conf` flag:

```toml
#Configuration.
log_level="info"

[redis]
addr="localhost:6379"
password=""
db=10

[mqtt]
server="tcp://broker-host:1883"
user="mqtt_user"
password="mqtt_password"

[gateway]
mac="b827ebfffe9448d0"

[band]
name="AU_915_928"

[device]
eui="0000000000000003"
address="019b58cf"
network_session_encription_key="13ef56f3089a68252cd7d873fcecf009"
serving_network_session_integrity_key="13ef56f3089a68252cd7d873fcecf009"
forwarding_network_session_integrity_key="13ef56f3089a68252cd7d873fcecf009"
application_session_key="9d12b80004300f957c154da245c68029"
marshaler="json"
nwk_key="00000000000000010000000000000001"
app_key="00000000000000010000000000000001"
join_eui="0000000000000003"
major=0
mac_version=1
mtype=2
profile="OTAA"
joined=false

[data_rate]
bandwith=125
spread_factor=10
bit_rate=0

[rx_info]
channel=0
code_rate="4/5"
crc_status=1
frequency=916800000
lora_snr=7.0
rf_chain=1
rssi=-57

[raw_payload]
payload="ff00"
use_raw=false

[[encoded_type]]
name="Flags"
is_float=false
num_bytes=1
value=5.0
max_value=255.0
min_value=0.0

[[encoded_type]]
name="Batería"
is_float=false
num_bytes=1
value=80.0
max_value=255.0
min_value=0.0

[[encoded_type]]
name="Luz"
is_float=false
num_bytes=1
value=50.0
max_value=255.0
min_value=-0.0
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

Values may be added using the `Add encoded type` button and setting the options:

![](images/new-gui.png?raw=true)

#### MAC Commands

All [lorawan package](https://github.com/brocaar/lorawan) end-device MAC commands are available to be sent with a message. Check desired mac commands and fill their payloads when needed.

### Building

The package is written in Go and tested with v 1.12. Make sure you have go installed before.  

The GUI is built using https://github.com/inkyblackness/imgui-go and OpenGL 3.2, so please check that repo to see requirements for your system. Once those are met, you may build the package like this: 

```
make
```

This will create the `gui` executable.