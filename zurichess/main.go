package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"runtime/pprof"
)

var (
	buildVersion = "glarus"
	buildTime    = "(just now)"

	cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
	version    = flag.Bool("version", false, "only print version and exit")
)

func init() {
	if buildTime == "(just now)" {
		// If build time is not known assume it is the modification time of the binary.
		path, err := exec.LookPath(os.Args[0])
		if err != nil {
			return
		}
		fi, err := os.Stat(path)
		if err != nil {
			return
		}
		buildTime = fi.ModTime().Format("2006-01-02 15:04:05")
	}
}

func main() {
	fmt.Printf("zurichess %v https://bitbucket.org/zurichess/zurichess\n", buildVersion)
	fmt.Printf("build with %v at %v, running on %v\n", runtime.Version(), buildTime, runtime.GOARCH)
	fmt.Printf("possible basis for zurirk racing kings engine\n")

	flag.Parse()
	if *version {
		return
	}
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	log.SetOutput(os.Stdout)
	log.SetPrefix("info string ")
	log.SetFlags(log.Lshortfile)

	uci := NewUCI()
	scan := bufio.NewScanner(os.Stdin)
	for scan.Scan() {
		line := scan.Text()
		if err := uci.Execute(line); err != nil {
			if err != errQuit {
				log.Println(err)
			} else {
				break
			}
		}
	}

	if scan.Err() != nil {
		log.Println(scan.Err())
	}
}
