package main

import (
	"log"
	"net"
	"net/http"
	"runtime"
	"strconv"
	"time"

	"github.com/codegangsta/negroni"
	"github.com/jaredfolkins/badactor"
	"github.com/julienschmidt/httprouter"
)

var st *badactor.Studio

var max int
var counter int

func main() {

	runtime.GOMAXPROCS(4)

	var dc int32
	var ac int32

	dc = 1024
	ac = 1024

	max = 10000
	counter = 0

	log.Printf("dc:ac:max:counter || %v:%v:%v:%v\n", dc, ac, max, counter)

	// init a new Studio
	st = badactor.NewStudio(dc)

	// add the rule to the stack
	ru := &badactor.Rule{
		Name:        "Login",
		Message:     "You have failed to login too many times",
		StrikeLimit: 3,
		ExpireBase:  time.Second * 60,
		Sentence:    time.Minute * 5,
	}
	st.AddRule(ru)

	err := st.CreateDirectors(ac)
	if err != nil {
		log.Fatal(err)
	}

	// start reaper
	st.StartReaper()

	// router
	router := httprouter.New()
	router.GET("/", IndexHandler)
	router.GET("/bench", BenchmarkInfractionWriteHandler)

	// middleware
	n := negroni.New(negroni.NewRecovery())
	n.Use(NewBadActorMiddleware())
	n.UseHandler(router)
	n.Run(":9999")

}

//
// MIDDLEWARE
//
type BadActorMiddleware struct {
	negroni.Handler
}

func NewBadActorMiddleware() *BadActorMiddleware {
	return &BadActorMiddleware{}
}

func (bam *BadActorMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {

	// snag the IP for use as the actor's name
	an, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		panic(err)
	}

	if st.IsJailed(an) {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	// call the next middleware in the chain
	next(w, r)
}

//
// HANDLER
//
func BenchmarkInfractionWriteHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// rule name
	rn := "Login"

	// actor name
	an := strconv.Itoa(counter)
	if counter > max {
		counter = 0
	} else {
		counter++
	}

	st.Infraction(an, rn)
	return
}

func IndexHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	return
}
