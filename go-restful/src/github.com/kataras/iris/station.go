// Copyright (c) 2016, Gerasimos Maropoulos
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without modification,
// are permitted provided that the following conditions are met:
//
// 1. Redistributions of source code must retain the above copyright notice,
//    this list of conditions and the following disclaimer.
//
// 2. Redistributions in binary form must reproduce the above copyright notice,
//	  this list of conditions and the following disclaimer
//    in the documentation and/or other materials provided with the distribution.
//
// 3. Neither the name of the copyright holder nor the names of its contributors may be used to endorse
//    or promote products derived from this software without specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
// ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
// WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
// DISCLAIMED. IN NO EVENT SHALL JULIEN SCHMIDT BE LIABLE FOR ANY
// DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
// (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
// LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND
// ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
// SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
package iris

import (
	"html/template"
	"net/http"
	"net/http/pprof"
	"os"
	"runtime"
	"sync"
	"time"
)

const (
	// DefaultProfilePath is the default path for the web pprof '/debug/pprof'
	DefaultProfilePath = "/debug/pprof"
)

type (
	// IStation is the interface which the Station should implements
	IStation interface {
		IRouter
		Serve() http.Handler
		Plugin(IPlugin) error
		GetPluginContainer() IPluginContainer
		GetTemplates() *template.Template
		//yes we need that again if no .Listen called and you use other server, you have to call .Build() before
		OptimusPrime()
		HasOptimized() bool
		GetLogger() *Logger
	}

	// StationOptions is the struct which contains all Iris' settings/options
	StationOptions struct {
		// Profile set to true to enable web pprof (debug profiling)
		// Default is false, enabling makes available these 7 routes:
		// /debug/pprof/cmdline
		// /debug/pprof/profile
		// /debug/pprof/symbol
		// /debug/pprof/goroutine
		// /debug/pprof/heap
		// /debug/pprof/threadcreate
		// /debug/pprof/pprof/block
		Profile bool

		// ProfilePath change it if you want other url path than the default
		// Default is /debug/pprof , which means yourhost.com/debug/pprof
		ProfilePath string

		// Cache for Router, change it to false if you don't want to use the cache mechanism that Iris provides for your routes
		Cache bool
		// CacheMaxItems max number of total cached routes, 500 = +~20000 bytes = ~0.019073MB
		// Every time the cache timer reach this number it will reset/clean itself
		// Default is 0
		// If <=0 then cache cleans all of items (bag)
		// Auto cache clean is happening after 5 minutes the last request serve, you can change this number by 'ResetDuration' property
		// Note that MaxItems doesn't means that the items never reach this lengh, only on timer tick this number is checked/consider.
		CacheMaxItems int
		// CacheResetDuration change this time.value to determinate how much duration after last request serving the cache must be reseted/cleaned
		// Default is 5 * time.Minute , Minimum is 30 seconds
		//
		// If CacheMaxItems <= 0 then it clears the whole cache bag at this duration.
		CacheResetDuration time.Duration

		// PathCorrection corrects and redirects the requested path to the registed path
		// for example, if /home/ path is requested but no handler for this Route found,
		// then the Router checks if /home handler exists, if yes, redirects the client to the correct path /home
		// and VICE - VERSA if /home/ is registed but /home is requested then it redirects to /home/
		//
		// Default is true
		PathCorrection bool
	}

	// Station is the container of all, server, router, cache and the sync.Pool
	Station struct {
		IRouter
		Server          *Server
		templates       *template.Template
		pool            sync.Pool
		options         StationOptions
		pluginContainer *PluginContainer
		//it's true if OptimusPrime has run one time
		optimized bool
		logger    *Logger
	}
)

// check at the compile time if Station implements correct the IRouter interface
// which it comes from the *Router or MemoryRouter which again it comes from *Router
var _ IStation = &Station{}

// newStation creates and returns a station, is used only inside main file iris.go
func newStation(options StationOptions) *Station {
	// create the station
	s := &Station{options: options, pluginContainer: &PluginContainer{}}
	// create the router
	var r IRouter
	//for now, we can't directly use NewRouter and after NewMemoryRouter, types are not the same.
	if options.Cache {
		r = NewMemoryRouter(NewRouter(s), options.CacheMaxItems, options.CacheResetDuration)
	} else {
		r = NewRouter(s)
	}

	// set the router
	s.IRouter = r

	// set the debug profiling handlers if enabled
	if options.Profile {
		debugPath := options.ProfilePath
		r.Get(debugPath+"/", ToHandlerFunc(pprof.Index))
		r.Get(debugPath+"/cmdline", ToHandlerFunc(pprof.Cmdline))
		r.Get(debugPath+"/profile", ToHandlerFunc(pprof.Profile))
		r.Get(debugPath+"/symbol", ToHandlerFunc(pprof.Symbol))

		r.Get(debugPath+"/goroutine", ToHandlerFunc(pprof.Handler("goroutine")))
		r.Get(debugPath+"/heap", ToHandlerFunc(pprof.Handler("heap")))
		r.Get(debugPath+"/threadcreate", ToHandlerFunc(pprof.Handler("threadcreate")))
		r.Get(debugPath+"/pprof/block", ToHandlerFunc(pprof.Handler("block")))
	}

	s.pool = sync.Pool{New: func() interface{} {
		return &Context{station: s, Params: make([]PathParameter, 0), mu: sync.Mutex{}}
	}}

	return s
}

// Plugin activates the plugins and if succeed then adds it to the activated plugins list
func (s *Station) Plugin(plugin IPlugin) error {
	return s.pluginContainer.Plugin(plugin)
}

// GetPluginContainer returns the pluginContainer
func (s Station) GetPluginContainer() IPluginContainer {
	return s.pluginContainer
}

// GetTemplates returns the *template.Template registed to this station, if any
func (s Station) GetTemplates() *template.Template {
	return s.templates
}

// GetLogger returns ( or creates if not exists already) the logger
func (s *Station) GetLogger() *Logger {
	if s.logger == nil {
		s.logger = NewLogger(LoggerOutTerminal, "", 0)
	}
	return s.logger
}

func (s *Station) forceOptimusPrime() {

	//check if any route has cors setted to true
	routerHasCors := func() bool {
		gLen := len(s.IRouter.getGarden())
		for i := 0; i < gLen; i++ {
			if s.IRouter.getGarden()[i].cors {
				return true
			}
		}
		return false
	}()

	if routerHasCors {
		s.IRouter.setMethodMatch(CorsMethodMatch)
	}

	// check if any route has subdomains
	routerHasHosts := func() bool {
		gLen := len(s.IRouter.getGarden())
		for i := 0; i < gLen; i++ {
			if s.IRouter.getGarden()[i].hosts {
				return true
			}
		}
		return false
	}()

	// For performance only,in order to not check at runtime for hosts and subdomains, I think it's better to do this:
	if routerHasHosts {
		switch s.IRouter.getType() {
		case Normal:
			s.IRouter = NewRouterDomain(s.IRouter.(*Router))
			break
		case Memory:
			s.IRouter = NewMemoryRouterDomain(s.IRouter.(*MemoryRouter))
			break
		}
		// just this no new router
	}

	//check for memoryrouter and use syncmemoryrouter & synccontextcache if cores > 1
	var r IMemoryRouter
	routerType := s.IRouter.getType()
	if routerType == Memory || routerType == DomainMemory {
		if routerType == Memory {
			r = s.IRouter.(*MemoryRouter)
		} else if routerType == DomainMemory {
			r = s.IRouter.(*MemoryRouterDomain)
		} else {
			panic("[Iris] From Station.OptimisPrime, unsupported Router, please post this as issue at github.com/kataras/iris")
		}

		if !r.hasCache() {
			var cache IContextCache

			cache = NewContextCache()
			//check if we have more than one core then use theMemoryRouterCache,otherwise use the SyncMemoryRouterCache and it's underline MemoryRouterCache
			if runtime.GOMAXPROCS(-1) > 1 {
				cache = NewSyncContextCache(cache.(*ContextCache))
				//println("SYNCED MULTI_CORE CACHE WITH CORES: ", runtime.GOMAXPROCS(-1))
			}

			r.setCache(cache)
		}

		// temp solution, wait for answer on this: https://github.com/kataras/iris/issues/44
		// wrk -t16 -c100 -d30s http://127.0.0.1:8080/rest/hello
		// no the problem is from net/http package:
		// https://github.com/golang/go/issues/14940
		// tries to modify the header after ServeHTTP
		// The problem is not Iris' router
		//if !r.hasCache() {
		//	cache := NewSyncMemoryRouterCache(NewMemoryRouterCache())
		//	r.setCache(cache)
		//}
		if runtime.GOMAXPROCS(-1) > 1 {
			s.IRouter = NewSyncRouter(r)
		}
	}
	s.optimized = true
}

// OptimusPrime make the best last optimizations to make iris the faster framework out there
// This function is called automatically on .Listen, but if you don't use .Listen or .Serve,
// then YOU MUST CALL .OptimusPrime before run a server
func (s *Station) OptimusPrime() {
	if !s.optimized {
		s.forceOptimusPrime()
	}

}

// HasOptimized returns if the station has optimized ( OptimusPrime run once)
func (s *Station) HasOptimized() bool {
	return s.optimized
}

// ServeHTTP returns the correct http.Handler for your machine and configuration, ready to use.
// it's a little hack I though in order to call OptimusPrime automatically,
// and without need of checking things every time a route added to the router
/*func (s *Station) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	s.once.Do(func() {
		//println("ServeHTTP: This ServeHTTP wrapper runs only once")
		s.OptimusPrime()
	})

	s.IRouter.ServeHTTP(res, req)
}*/

// Listen starts the standalone http server
// which listens to the fullHostOrPort parameter which as the form of
// host:port or just port
func (s *Station) Listen(fullHostOrPort ...string) error {
	s.OptimusPrime()

	s.pluginContainer.DoPreListen(s)
	// I moved the s.Server here because we want to be able to change the Router before listen (with plugins)
	// set the server with the server handler
	s.Server = &Server{handler: s.IRouter}
	err := s.Server.listen(fullHostOrPort...)
	if err == nil {
		s.pluginContainer.DoPostListen(s)
		ch := make(chan os.Signal)
		<-ch
		s.Close()
	}

	return err
}

// ListenTLS Starts a httpS/http2 server with certificates,
// if you use this method the requests of the form of 'http://' will fail
// only https:// connections are allowed
// which listens to the fullHostOrPort parameter which as the form of
// host:port or just port
func (s *Station) ListenTLS(fullAddress string, certFile, keyFile string) error {
	s.OptimusPrime()
	s.pluginContainer.DoPreListen(s)
	// I moved the s.Server here because we want to be able to change the Router before listen (with plugins)
	// set the server with the server handler
	s.Server = &Server{handler: s.IRouter}
	err := s.Server.listenTLS(fullAddress, certFile, keyFile)
	if err == nil {
		s.pluginContainer.DoPostListen(s)
		ch := make(chan os.Signal)
		<-ch
		s.Close()
	}

	return err
}

// Serve is used instead of the iris.Listen
// eg  http.ListenAndServe(":80",iris.Serve()) if you don't want to use iris.Listen(":80")
func (s *Station) Serve() http.Handler {
	s.OptimusPrime()
	s.pluginContainer.DoPreListen(s)
	return s.IRouter
}

// Close is used to close the tcp listener from the server
func (s *Station) Close() {
	s.pluginContainer.DoPreClose(s)
	s.Server.closeServer()

}

// Templates sets the templates glob path for the web app
func (s *Station) Templates(pathGlob string) {
	var err error
	//s.htmlTemplates = template.Must(template.ParseGlob(pathGlob))
	s.templates, err = template.ParseGlob(pathGlob)

	if err != nil {
		//if err then try to load the same path but with the current directory prefix
		// and if not success again then just panic with the first error
		pwd, cerr := os.Getwd()
		if cerr != nil {
			panic(err.Error())

		}
		s.templates, cerr = template.ParseGlob(pwd + pathGlob)
		if cerr != nil {
			panic(err.Error())
		}
	}

}
