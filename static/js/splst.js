$(function() {

    $('.closeit').click(function() {
        $(this).parent().fadeOut("fast");
    });

    $('.add-project').click(function() {
        if ($('#add-project-link').attr('disabled')) {
            // We are in the midle of clearing the form
            return;
        }

        $('#add-project-container').slideToggle('fast');

        if (!$('#fetch').is(':hidden') && !$('#add-project-link').hasClass("loading")) {
            $('#url').removeAttr("disabled");
            $('#fetch').removeAttr("disabled");
        }

        if ($('#fetch').is(':hidden')) {
            $('#name').focus();
        } else {
            $('#url').focus();
        }
    });

    var _cache = {};

    $('#fetch').click(function () {

        url = $.trim($("#url").val());

        if (url == '') {
            return false;
        }

        if (url.substr(0, 7) != 'http://' && url.substr(0, 8) != 'https://') {
            url = 'http://' + url;
            $("#url").val(url);
        }

        $('#add-project-link').addClass("loading");
        $("#url").attr("disabled", "disabled");
        $("#fetch").attr("disabled", "disabled");

        failedToFetch = function() {
            $("#add-project-link").removeClass("loading");
            $('#fetch').removeAttr("disabled");

            url = $('#url').val();
            $("#add-project").find('input:text, textarea').val('');
            $('#url').removeAttr("disabled").val(url);

            group = $('#url-group').addClass("error");
            _cache['url'] = $(".help-inline", group).text();
            $(".help-inline", group).text('Could not fetch the url');
            $('#url').focus();

            $('button[type="submit"]', $("#add-project")).removeAttr("disabled");
        }

        function fixedEncodeURI (str) {
                return encodeURI(str).replace(/%5B/g, '[').replace(/%5D/g, ']');
        }

        $.ajax({
            url: "/url-info",
            data: {url: fixedEncodeURI(url)},
            type: "GET",
            dataType: 'json',
            accepts: {
                json: 'application/json'
            }
        }).always(function(result) {
            $.each(_cache, function(field, inline_help) {
                group = $("#"+field+"-group");
                group.removeClass("error");
                $(".help-inline", group).text(inline_help);
            });
            // clear the cache
            _cache = {};

            return false;

        }).fail(failedToFetch)
          .success(function(result) {
            if (result) {
                $("#add-project-link").removeClass("loading");
                $('#fetch').removeAttr("disabled");

                url = $('#url').val();
                $("#add-project").find('input:text, textarea').val('');
                $('#url').removeAttr("disabled").val(url);

                $.each(result, function(field, value) {
                    $('#'+field).val(value);
                });
                $('.second-stage').slideDown('fast');
                setTimeout(function() {
                    $('.first-stage').slideUp('fast');
                }, 500);
                $('#fetch').hide();
                $('#close').hide();
                $('#add').show();
                $('#clear').show();

                $('button[type="submit"]', $("#add-project")).removeAttr("disabled");

                return true;
            }

            failedToFetch();
            return false;
        });
    });

    clear = function() {
        $('#add-project-container').slideUp('fast');
        setTimeout(function() {
            // We need to clear text input
            // $("#add-form").find('input:text, input:password, input:file, select, textarea').val('');
            // $("#add-form").find('input:radio, input:checkbox').removeAttr('checked').removeAttr('selected');
            $("#add-project").find('input:text, textarea').val('');
            $('.first-stage').show();
            $('.second-stage').hide();
            $('button[type="submit"]', $("#add-project")).removeAttr("disabled");
            $('#add-project-link').removeAttr("disabled");
            $("#url").removeAttr("disabled");
            $('#fetch').show();
            $('#close').show();
            $('#add').hide();
            $('#clear').hide();
        }, 500);
    }

    $('#add-project').submit(function() {

        $('button[type="submit"]', $(this)).attr("disabled", "disabled");
        $('#add-project-link').addClass("loading");
        $('#error-block').fadeOut("slow");

        $.ajax({
            url: "/project",
            data: $(this).serializeArray(),
            type: "POST",
            dataType: 'json',
            accepts: {
                json: 'application/json'
            }
        }).always(function(result) {
            $.each(_cache, function(field, inline_help) {
                group = $("#"+field+"-group");
                group.removeClass("error");
                $(".help-inline", group).text(inline_help);
            });
            // clear the cache
            _cache = {};

            return false;

        }).fail(function(result) {

            $("#add-project-link").removeClass("loading");

            if (result.getResponseHeader("Content-Type") == "application/json") {
                errMessage = $.parseJSON(result.responseText);
                $.each(errMessage, function(field, error) {

                    if (field == "error") {
                        $("#error-message").text(error);
                        $("#error-block").fadeIn("fast");
                    } else {
                        group = $("#"+field+"-group");
                        // save inline help message in the cache
                        _cache[field] = $(".help-inline", group).text();
                        group.addClass("error");
                        $(".help-inline", group).text(error);
                    }
                });
            }

            $('button[type="submit"]', $("#add-project")).removeAttr("disabled");

        }).success(function(result) {

            $("#add-project-link").removeClass("loading");
            $('#add-project-link').attr("disabled", "disabled");
            $('#upload-message').fadeIn("slow");
            setTimeout(function() {
                $('#upload-message').fadeOut("slow");

                // Increase jobs count and show it
                count = parseInt($('small', $('#jobs-count')).text());
                $('small', $('#jobs-count')).text(++count);
                $('#jobs-count').show();

                setTimeout(clear, 1000);
            }, 2000);
        });

        return false;
    });

    $('#clear').click(clear);

    $('.toggle-delete').click(function () {
        var pid = $(this).attr('data-pid');
        $('#'+pid).slideToggle('fast');
        $('#error-'+pid).slideUp('fast');
    });

    $('.delete').click(function () {
        var pid = $(this).attr('data-pid');
        $.ajax({
            url: "/project/"+pid,
            type: "DELETE",
            dataType: 'json',
            accepts: {
                json: 'application/json'
            }
        }).fail(function(result) {
            $('#'+pid).slideToggle('fast');
            $('#error-'+pid).slideDown('fast');
        }).success(function(result) {
            $('#container-'+pid).fadeToggle('slow');
            setTimeout(function() {
                $('#container-'+pid).remove();

                // Decrease my-projects count
                count = parseInt($('small', $('#my-projects-count')).text());
                $('small', $('#my-projects-count')).text(--count);
            }, 500);
        });
    });

    $('.thumbnail').hover(function () {
        // In
        $('.actions', this).slideDown('fast');
    }, function () {
        // Out
        $('.actions', this).slideUp('fast');
        $('.alert', this).slideUp('fast');
    });

    $('#my-projects-count').tooltip({placement: 'top', title: function () {
        count = $('small', this).text();
        if (count == 0) {
            return "You haven't submitted any projects yet!";
        } else if (count == 1) {
            return "Great! You already have a project";
        } else {
            return count+" projects. Good job!";
        }
    }});

    $('#jobs-count').tooltip({placement: 'top', title: function () {
        count = $('small', this).text();
        if (count == 1) {
            return "You have "+count+" job in the queue";
        } else if (count > 1) {
            return "You have "+count+" jobs in the queue";
        }
    }});
});
