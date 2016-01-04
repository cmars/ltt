package main

import (
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/julienschmidt/httprouter"
)

func defaultPath() string {
	home := os.Getenv("HOME")
	if home == "" {
		log.Fatal("HOME environment variable not set")
	}
	return filepath.Join(home, "Music", "listentothis")
}

func init() {
	rand.Seed(time.Now().Unix())
}

type service struct {
	path string
}

func newService() (*service, error) {
	s := &service{path: defaultPath()}
	err := os.MkdirAll(s.keepDir(), 0755)
	if err != nil {
		return nil, err
	}
	err = os.MkdirAll(s.trashDir(), 0755)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (s *service) keepDir() string {
	return filepath.Join(s.path, "Keep")
}

func (s *service) trashDir() string {
	return filepath.Join(s.path, "Trash")
}

func main() {
	s, err := newService()
	if err != nil {
		log.Fatal(err)
	}

	r := httprouter.New()
	r.GET("/", s.index)
	r.GET("/song/:filename", s.index)
	r.POST("/song/:filename", s.keep)
	r.DELETE("/song/:filename", s.trash)
	r.ServeFiles("/files/*filepath", http.Dir(s.path))

	log.Fatal(http.ListenAndServe("127.0.0.1:8080", r))
}

func (*service) index(w http.ResponseWriter, req *http.Request, p httprouter.Params) {
	var err error
	filename := p.ByName("filename")
	if filename == "" {
		filename, err = randomFilename()
		if err == ErrNotFound {
			// TODO: react gracefully here
			http.Error(w, "no files found", http.StatusNotFound)
			return
		} else if err != nil {
			http.Error(w, fmt.Sprintf("failed to select random filename: %v", err), http.StatusInternalServerError)
			return
		}
	}

	err = playTemplate.Execute(w, struct {
		Filename string
	}{
		Filename: filename,
	})
	if err != nil {
		http.Error(w, "failed to execute template", http.StatusInternalServerError)
	}
}

var ErrNotFound = fmt.Errorf("not found")

func randomFilename() (string, error) {
	matches, err := filepath.Glob(filepath.Join(defaultPath(), "*.ogg"))
	if err != nil {
		return "", err
	}
	if len(matches) == 0 {
		return "", ErrNotFound
	}
	n := rand.Intn(len(matches))
	return filepath.Base(matches[n]), nil
}

func (s *service) keep(w http.ResponseWriter, req *http.Request, p httprouter.Params) {
	filename := p.ByName("filename")
	if filename == "" {
		http.Error(w, "missing required filename parameter", http.StatusBadRequest)
		return
	}

	err := s.moveFile(filename, s.keepDir())
	if err != nil {
		log.Println(err)
		http.Error(w, fmt.Sprintf("failed to rename song file: %v", err), http.StatusInternalServerError)
		return
	}
}

func (s *service) trash(w http.ResponseWriter, req *http.Request, p httprouter.Params) {
	filename := p.ByName("filename")
	if filename == "" {
		http.Error(w, "missing required filename parameter", http.StatusBadRequest)
		return
	}

	err := s.moveFile(filename, s.trashDir())
	if err != nil {
		log.Println(err)
		http.Error(w, fmt.Sprintf("failed to rename song file: %v", err), http.StatusInternalServerError)
		return
	}
}

func (s *service) moveFile(filename, newPath string) error {
	oldpath := filepath.Join(s.path, filename)
	newpath := filepath.Join(newPath, filename)
	return os.Rename(oldpath, newpath)
}

// TODO: make this all fancy and shit
var playTemplate = template.Must(template.New("play").Parse(`<html>
<head>

<link href="https://cdnjs.cloudflare.com/ajax/libs/jplayer/2.9.2/skin/blue.monday/css/jplayer.blue.monday.min.css" rel="stylesheet" type="text/css" />
<script type="text/javascript" src="https://cdnjs.cloudflare.com/ajax/libs/jquery/3.0.0-alpha1/jquery.min.js"></script>
<script type="text/javascript" src="https://cdnjs.cloudflare.com/ajax/libs/jplayer/2.9.2/jplayer/jquery.jplayer.min.js"></script>
<script type="text/javascript" src="https://cdnjs.cloudflare.com/ajax/libs/jplayer/2.9.2/add-on/jquery.jplayer.inspector.min.js"></script>

</head>
<body>
<h1>{{ .Filename }}</h1>

<script type="text/javascript">
//<![CDATA[

next = function(){
	window.location.href = "/";
};

keep = function(){
	$.ajax({
		type: "POST",
		url: "/song/{{ .Filename }}",
	}).done(function(){
		next();
	})
};

trash = function(){
	$.ajax({
		type: "DELETE",
		url: "/song/{{ .Filename }}",
	}).done(function(){
		next();
	})
};

$(document).ready(function(){

	$("#jquery_jplayer_1").jPlayer({
		ready: function (event) {
			$(this).jPlayer("setMedia", {
				title: "{{ .Filename }}",
				oga: "/files/{{ .Filename }}"
			}).jPlayer("play");
		},
		ended: function (event) {
			window.location.href = "/";
		},
		supplied: "oga",
		wmode: "window",
		useStateClassSkin: true,
		autoBlur: false,
		smoothPlayBar: true,
		keyEnabled: true,
		remainingDuration: true,
		toggleDuration: true
	});

});

//]]>
</script>

<a href="javascript:keep();">Keep</a>

<div id="jquery_jplayer_1" class="jp-jplayer"></div>
<div id="jp_container_1" class="jp-audio" role="application" aria-label="media player">
	<div class="jp-type-single">
		<div class="jp-gui jp-interface">
			<div class="jp-volume-controls">
				<button class="jp-mute" role="button" tabindex="0">mute</button>
				<button class="jp-volume-max" role="button" tabindex="0">max volume</button>
				<div class="jp-volume-bar">
					<div class="jp-volume-bar-value"></div>
				</div>
			</div>
			<div class="jp-controls-holder">
				<div class="jp-controls">
					<button class="jp-play" role="button" tabindex="0">play</button>
					<button class="jp-stop" role="button" tabindex="0">stop</button>
				</div>
				<div class="jp-progress">
					<div class="jp-seek-bar">
						<div class="jp-play-bar"></div>
					</div>
				</div>
				<div class="jp-current-time" role="timer" aria-label="time">&nbsp;</div>
				<div class="jp-duration" role="timer" aria-label="duration">&nbsp;</div>
				<div class="jp-toggles">
					<button class="jp-repeat" role="button" tabindex="0">repeat</button>
				</div>
			</div>
		</div>
		<div class="jp-details">
			<div class="jp-title" aria-label="title">&nbsp;</div>
		</div>
		<div class="jp-no-solution">
			<span>Update Required</span>
			To play the media you will need to either update your browser to a recent version or update your <a href="http://get.adobe.com/flashplayer/" target="_blank">Flash plugin</a>.
		</div>
	</div>
</div>

<a href="javascript:next();">Next</a>

<a href="javascript:trash();">Trash</a>

<!-- <div id="jplayer_inspector"></div> -->

</body>
</html>
`))
