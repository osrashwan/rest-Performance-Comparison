package com.colobu.rest.rapidoid;

import org.rapidoid.http.fast.*;
  // wget https://repo.maven.apache.org/maven2/org/rapidoid/rapidoid-http-fast/5.0.7/rapidoid-http-fast-5.0.7.jar
  public class Main 
  {
    public static void main (String[] args)
    {
    	On.port(8080);
    	On.get ("/rest/hello").json("Hello World!!");
    }
  }
