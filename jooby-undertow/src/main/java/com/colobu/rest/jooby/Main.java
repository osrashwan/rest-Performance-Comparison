package com.colobu.rest.jooby;




import org.jooby.Jooby;

public class Main extends Jooby { // 1

  {
    // 2
    get("/rest/hello", () -> "Hello World!");
  }

  public static void main(final String[] args) throws Exception {
    run(Main::new, args); // 3. start the application.
  }

}