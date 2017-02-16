package main

import (
	"C"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/ry/v8worker"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"regexp"
	"runtime"
)

type module struct {
	Err      error  `json:"err"`
	Source   string `json:"source"`
	Id       string `json:"id"`
	Filename string `json:"filename"`
	Dirname  string `json:"dirname"`
	main     bool
}

var (
	jsExtensionRe = regexp.MustCompile(`\.js$`)
	jsFile        = flag.String("f", "server.js", "js file to run")
)

// Adapted from node.js source:
// see https://github.com/nodejs/node/blob/master/src/node.js#L871
const nativeModule = `
	'use strict';

	function NativeModule(rawModule) {
		this.filename = rawModule.filename;
		this.dirname = rawModule.dirname;
		this.id = rawModule.id;
		this.exports = {};
		this.loaded = false;
		this._source = rawModule.source;
	}

	NativeModule.require = function(id) {
                console.log("ID:", id)
		var rawModule = JSON.parse($sendSync(id));
		if (rawModule.err) {
			throw new RangeError(JSON.stringify(rawModule.err));
		}

		var nativeModule = new NativeModule(rawModule);

		nativeModule.compile();

		return nativeModule.exports;
	};

	NativeModule.prototype.compile = function() {
		var fn = eval(this._source);
		fn(this.exports, NativeModule.require, this, this.filename, this.dirname);
		this.loaded = true;
	};
	`

func (m *module) load() {
	filename := jsExtensionRe.ReplaceAllString(m.Id, "") + ".js"
	if wd, err := os.Getwd(); err == nil {
		m.Filename = path.Join(wd, filename)
	} else {
		m.Err = err
		return
	}
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		m.Err = err
		return
	}
	m.Dirname = path.Dir(m.Filename)
	var b bytes.Buffer
	if m.main {
		b.WriteString(fmt.Sprintf(
			"var main = new NativeModule({ id: '%s', filename: '%s', dirname: '%s' });\n",
			m.Id, m.Filename, m.Dirname))
	}
	b.WriteString("(function (exports, require, module, __filename, __dirname) { ")
	if m.main {
		b.WriteString("\nrequire.main = module;")
	}
	b.Write(file)
	if m.main {
		b.WriteString("\n}")
		b.WriteString("(main.exports, NativeModule.require, main, main.filename, main.dirname));")
		b.WriteString("\n$send('exit');") // exit when main returns
	} else {
		b.WriteString("\n});")
	}
	m.Source = b.String()
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	worker := v8worker.New(func(msg string) {
		log.Printf(msg)
		w.Write([]byte(msg))
	}, func(msg string) string {
		m := module{Id: msg, main: false}
		m.load()
		bytes, _ := json.Marshal(m)
		return string(bytes)
	})

	defer func() {
		worker.TerminateExecution()
	}()

	if err := worker.Load("native-module.js", nativeModule); err != nil {
		log.Println(err)
		return
	}
	JSCode := `
        $send("Hello world from V8\n");
    `
	if err := worker.Load("code.js", JSCode); err != nil {
		log.Printf("failed to load js file. error: %v", err)
	}
}

func main() {
	fmt.Println("Go version:", runtime.Version())

	http.HandleFunc("/", rootHandler)
	http.ListenAndServe(":8080", nil)
}

func loadMainModule(w *v8worker.Worker, id string) error {
	m := module{Id: id, main: true}
	m.load()
	if m.Err != nil {
		return m.Err
	}
	return w.Load(m.Filename, m.Source)
}
