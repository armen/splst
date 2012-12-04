BUILD=`date +'%Y%m%d%H%M%S'`

install:
	go get splst

styles:
	echo ${BUILD} > conf/BUILD
	git clean -fX static/css
	lessc -x --yui-compress less/bootstrap.less static/css/bootstrap.${BUILD}.css
	lessc -x --yui-compress less/responsive.less static/css/bootstrap-responsive.${BUILD}.css
	lessc -x --yui-compress less/splst.less static/css/splst.${BUILD}.css

uglify:
	uglifyjs static/js/jquery-1.7.2.js > static/js/min/jquery-1.7.2.min.js
	uglifyjs static/js/splst.js > static/js/min/splst.min.js
	uglifyjs static/js/bootstrap-tooltip.js > static/js/min/bootstrap-tooltip.min.js
