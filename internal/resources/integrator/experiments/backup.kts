# Look with trigger

rest()
.post("/")
.to("direct:mqtt-response-handler");

from("direct:mqtt-response-handler")
.to("log:info")
.to("paho:mytopic-response?brokerUrl=tcp://mqtt-broker:1883");



# Work manual, but not with trigger
rest {
    path("/") {
        post("/") {
            to("direct:mqtt-response-handler")
        }
    }
}

from("direct:mqtt-response-handler")
.to("log:info")
.to("paho:mytopic-response?brokerUrl=tcp://mqtt-broker:1883")
.process { e -> e.getIn().setBody("") };
