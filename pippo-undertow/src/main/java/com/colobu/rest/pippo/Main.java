package com.colobu.rest.pippo;

import ro.pippo.core.Pippo;

public class Main {
    
	public static void main(String[] args) {
        Pippo pippo = new Pippo();
        pippo.getServer().getSettings().port(8080);
        pippo.getServer().getSettings().host("0.0.0.0");
        pippo.getApplication().GET("/rest/hello", (routeContext) -> routeContext.send("Hello World!"));
        pippo.start();
    }
}
