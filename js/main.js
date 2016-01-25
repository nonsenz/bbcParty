var bbcParty = (function main() {
    var hit = "";

    nanoajax.ajax({url:'/hit'}, function (code, responseText) {
        hit = JSON.parse(responseText).id;
    });

    return {
        nextHit: function() {return hit;}
    };
})();

var tag = document.createElement('script');

tag.src = "https://www.youtube.com/iframe_api";
var firstScriptTag = document.getElementsByTagName('script')[0];
firstScriptTag.parentNode.insertBefore(tag, firstScriptTag);

var player;
function onYouTubeIframeAPIReady() {

    if (!bbcParty.nextHit().length) {
        setTimeout(onYouTubeIframeAPIReady, 100);
        return;
    }

    player = new YT.Player('player', {
        height: '390',
        width: '640',
        events: {
            'onReady': onPlayerReady
        }
    });

    console.log("data", player);
}

// 4. The API will call this function when the video player is ready.
function onPlayerReady(event) {
    player.loadVideoById(bbcParty.nextHit());
    event.target.playVideo();
}