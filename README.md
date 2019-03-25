## Loraserver device simulator

This is an utility program to simulate devices for the [loraserver](https://loraserver.io) project.  Basically, it acts as a `lora-gateway-bridge` middleman, publishing and receiving messages through MQTT.
It supports all bands and configurations LoRaWAN versions 1.0 and 1.1.  

It has a simple but complete GUI built with [imgui-go](https://github.com/inkyblackness/imgui-go) and OpenGL 3.2, that allows to configure everything that's needed, such as MQTT broker and credentials, device keys, LoRaWAN version, message marshaling method, data payload, etc.

**Important**: This is a work in progress: the `cli` mode needs to be rewritten (it doesn't work right now) and there may be bugs in the `gui` version. Please report any by filing and issue.

![](images/new-gui.png?raw=true)

### Requirements

As mentioned, this program needs OpenGL 3.2 to be installed. Also, it uses Redis to store device addres and keys, frame counters and nonces in OTAA mode.

### Conf

The GUI allows to modify all options, but they may also be seeded with a conf file for ease of use. An example file is provided to get an idea, but the program will only load a conf file with the name `conf.toml` located at the same dir as the binary, or if the path is given with the `--conf` flag:

```toml
#Configuration.
log_level = "info"

[mqtt]
  server = "tcp://localhost:1883"
  user = "username"
  password = "password"

[gateway]
  mac = "b827ebfffe9448d0"

[band]
  name = "AU_915_928"

[Device]
	eui="0000000000000000"
	address="000f6e3b"
	network_session_encription_key="dc5351f56794ed3ac17c382927192858"
	serving_network_session_integrity_key="dc5351f56794ed3ac17c382927192858"
	forwarding_network_session_integrity_key="dc5351f56794ed3ac17c382927192858"
	application_session_key="7b14565ba0e30d6ced804393fd6a0dd5"
	marshaler="json"
	nwk_key="00000000000000010000000000000001"
	app_key="00000000000000010000000000000001"
	join_eui="0000000000000002"
	mac_version=1
	profile="OTAA"
	joined=false
	skip_fcnt_check=true

[data_rate]
  bandwith = 125
  spread_factor = 10
  bit_rate = 0

[rx_info]
  channel = 0
  code_rate = "4/5"
  crc_status = 1
  frequency = 916800000
  lora_snr = 7.0
  rf_chain = 1
  rssi = -57

[raw_payload]
  payload = "ff00"
  use_raw = false
	script = "\n// Encode encodes the given object into an array of bytes.\n//  - fPort contains the LoRaWAN fPort number\n//  - obj is an object, e.g. {\"temperature\": 22.5}\n// The function must return an array of bytes, e.g. [225, 230, 255, 0]\nfunction Encode(fPort, obj) {\n\treturn [\n      obj[\"Flags\"],\n      obj[\"Battery\"],\n      obj[\"Light\"],\n    ];\n}\n"
  use_encoder = true
  max_exec_time = 500
  js_object = "{\n \"Flags\": 0,\n \"Battery\": 65,\n \"Light\": 54\n}"
  fport = 2

[[encoded_type]]
  name = "Flags"
  value = 5.0
  max_value = 255.0
  min_value = 0.0
  is_float = false
  num_bytes = 1

[[encoded_type]]
  name = "Batería"
  value = 80.0
  max_value = 255.0
  min_value = 0.0
  is_float = false
  num_bytes = 1

[[encoded_type]]
  name = "Luz"
  value = 50.0
  max_value = 255.0
  min_value = -0.0
  is_float = false
  num_bytes = 1

[redis]
  addr = "localhost:6379"
  password = ""
  db = 10
```
You may also import files located at `working-dir/confs` and save to the same directory.

When OTAA is set and the device is joined, uponinitialization the program will try to load keys and relevant data from Redis, overriding keys from the file.

### Data

The data to be sent may be presented as a hex string representation of the raw bytes, using a JS object and a decoding function to extract a bytes array from it, or using our encoding method (which then needs to be decoded accordingly at `lora-app-server`). As a reference, this is how we encode our data:

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

When using our encoding method, values may be added using the `Add encoded type` button and setting the options.  

To use your own custom JS encoder, click the "Use encoder" checkbox and the "Open decoder" button to open the form:

![](images/encoder.png?raw=true)

#### MAC Commands

All [lorawan package](https://github.com/brocaar/lorawan) end-device MAC commands are available to be sent with a message. Check desired mac commands and fill their payloads when needed.

### Building

The package is written in Go and tested with v 1.12. Make sure you have go installed before.  

The GUI is built using https://github.com/inkyblackness/imgui-go and OpenGL 3.2, so please check that repo to see requirements for your system. Once those are met, you may build the package like this: 

```
make
```

This will create the `gui` executable.