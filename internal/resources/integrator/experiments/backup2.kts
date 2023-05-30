
rest()
.post("/")
.to("direct:mqtt-response-handler");

from("direct:mqtt-response-handler")
.to("log:info?showAll=true&multiline=true")
.to("paho:mytopic-response?brokerUrl=tcp://mqtt-broker:1883")
.process { e -> e.getIn().setBody("") };