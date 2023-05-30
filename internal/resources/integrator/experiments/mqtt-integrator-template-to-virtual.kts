from("paho:mytopic_0?brokerUrl=tcp://mqtt-broker:1883")
.to("log:info")
.setHeader("ce-type").constant("CE-Type")
.setHeader("ce-source").constant("CE-Source")
.to("knative:endpoint/edge-service");