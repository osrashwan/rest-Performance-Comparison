package com.colobu.rest.ratpack;

import java.net.InetAddress;
import java.net.URI;

import ratpack.server.RatpackServer;
import ratpack.server.ServerConfig;

public class Main {
    
    	public static void main(String... args) throws Exception {
    		InetAddress addr = InetAddress.getByName("0.0.0.0");
    		   RatpackServer.start(server -> server
    		     .serverConfig(ServerConfig.embedded().address(addr).port(8080))
    		     .registryOf(registry -> registry.add("World!"))
    		     .handlers(chain -> chain
    		       .get("rest/hello",ctx -> ctx.render("Hello " + ctx.get(String.class)))
    		       .get(":name", ctx -> ctx.render("Hello " + ctx.getPathTokens().get("name") + "!"))     
    		     )
    		   );
    		 }
}
