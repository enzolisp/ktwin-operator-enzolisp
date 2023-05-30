from("timer:tick")
    .process { e -> e.getIn().body = "Hello Camel K!" }
    .to("log:info")