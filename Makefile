install:
	lessc -x --yui-compress less/bootstrap.less static/css/bootstrap.css
	lessc -x --yui-compress less/responsive.less static/css/bootstrap-responsive.css
	lessc -x --yui-compress less/splst.less static/css/splst.css
	go get splst

uglify:
	uglifyjs static/js/jquery-1.7.2.js > static/js/min/jquery-1.7.2.min.js
