function getCookieValue(cookieName) {
    var cookieArray = document.cookie.split(';');
    for (var i = 0; i < cookieArray.length; i++) {
        var cookie = cookieArray[i].trim();
        if (cookie.indexOf(cookieName + '=') == 0) {
            return cookie.substring(cookieName.length + 1, cookie.length);
        }
    }
    return null;
};

function getUsername() {
    return getCookieValue('username')
};


function getUrlParam(key) {
    var locationParams = new URLSearchParams(location.search);
    var val = locationParams.get(key);
    return decodeURIComponent(val);
};