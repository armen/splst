$(function() {

    $('#signin').popover({
        html: true,
        placement: "bottom",
        trigger: "hover",
        title: "Signin & Signup",
        content: '<form action="/google-signin" method="POST"><div class="controls"><button type="submit" class="btn btn-google btn-small">'+
                 '<i class="icon-google-plus icon-large icon-separator"></i>&nbsp;&nbsp;Signin with Google</button></div></form>',
        template: '<div class="popover"><div class="arrow"></div><div class="popover-inner"><h3 class="popover-title"></h3><div class="popover-content"><span></span></div></div></div>',
        delay: { show: 50, hide: 3000 },
    });
});
