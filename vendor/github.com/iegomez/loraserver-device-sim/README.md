# loraserver-device-sim
This package aims to help test the [loraserver](https://loraserver.io) infrastructure.  

It allows to "simulate" a device by publishing an encrypted message directly to the loraserver as if it came from lora-gateway-bridge.

It's a work in progress, so now it only allows for LoRaWAN 1.0.X specification messages (tested creating an ABP device with version 1.0.2 and regional parameters B). Also, only an Uplink and a test Join message may be sent (there´s no way to check for downlinks yet). So at the oment it's probably useful just to send uplinks to an ABP device with relaxed frame counter (though you can set the frame counter and it will increase on unconfirmed data up).

## Examples

You can check the examples folder for an example of an Uplink message. More will be added as new features are implemented.

