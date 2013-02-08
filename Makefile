BUILD = $(shell date +'%Y%m%d%H%M%S')

install:
	go get splst

styles:
	echo ${BUILD} > conf/BUILD
	git clean -fX static/css
	lessc -x --yui-compress less/bootstrap.less static/css/bootstrap.${BUILD}.css
	lessc -x --yui-compress less/responsive.less static/css/bootstrap-responsive.${BUILD}.css
	lessc -x --yui-compress less/gapp.less static/css/gapp.${BUILD}.css
	lessc -x --yui-compress less/splst.less static/css/splst.${BUILD}.css

uglify:
	uglifyjs static/js/jquery-1.7.2.js > static/js/min/jquery-1.7.2.min.js
	uglifyjs static/js/gapp.js > static/js/min/gapp.min.js
	uglifyjs static/js/splst.js > static/js/min/splst.min.js
	uglifyjs static/js/bootstrap-tooltip.js > static/js/min/bootstrap-tooltip.min.js
	uglifyjs static/js/bootstrap-popover.js > static/js/min/bootstrap-popover.min.js
	uglifyjs static/js/bootstrap-tagmanager.js > static/js/min/bootstrap-tagmanager.min.js

gofmt-hook:
	wget -q https://raw.github.com/edsrzf/gofmt-git-hook/master/fmt-check -O /tmp/pre-commit
	mv --interactive /tmp/pre-commit $(shell git rev-parse --git-dir)/hooks/
	chmod +x $(shell git rev-parse --git-dir)/hooks/pre-commit
