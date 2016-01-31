var bbcParty = (function main() {
        var hit = "",
            getNextHit = function() {
                nanoajax.ajax({url:'/hit'}, function (code, responseText) {
                    hit = JSON.parse(responseText).id;
                });
            },
            getBodySize = function() {
                var w = window,
                    d = document,
                    e = d.documentElement,
                    g = d.body,
                    x = w.innerWidth || e.clientWidth || g.clientWidth,
                    y = w.innerHeight|| e.clientHeight|| g.clientHeight;
                return {'x': x, 'y': y};
            };

        getNextHit();

        return {
            getBodySize: getBodySize,
            nextHit: function() {getNextHit(); return hit;}
        };
    })(),
    tag = document.createElement('script'),
    firstScriptTag = document.getElementsByTagName('script')[0],
    player;

tag.src = "https://www.youtube.com/iframe_api";
firstScriptTag.parentNode.insertBefore(tag, firstScriptTag);

function onYouTubeIframeAPIReady() {

    if (!bbcParty.nextHit().length) {
        setTimeout(onYouTubeIframeAPIReady, 100);
        return;
    }

    bodySize = bbcParty.getBodySize();
    player = new YT.Player('player', {
        height: bodySize.y - 20,
        width: bodySize.x - 20,
        events: {
            'onReady': onPlayerReady,
            'onStateChange': onStateChange
        }
    });
}

function onPlayerReady(event) {
    player.loadVideoById(bbcParty.nextHit());
    event.target.playVideo();
}

function onStateChange(event) {
    if (event.data == YT.PlayerState.ENDED) {
        player.loadVideoById(bbcParty.nextHit());
    }
}

window.onresize = function(event) {
    bodySize = bbcParty.getBodySize();
    player.setSize(bodySize.x - 20 , bodySize.y - 20);
};

document.onkeydown = function (e) {
    e = e || window.event;

    if (e.keyCode == '39') {
        player.loadVideoById(bbcParty.nextHit());
    }
};