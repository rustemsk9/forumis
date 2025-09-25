// Debug function to manually check cookie
function checkCookieInBrowser() {
    console.log("All cookies:", document.cookie);
    console.log("Looking for _cookie...");
    
    const cookies = document.cookie.split(';');
    let foundCookie = false;
    
    cookies.forEach(cookie => {
        const [name, value] = cookie.trim().split('&');
        console.log(`Cookie: ${name} = ${value}`);
        if (name === '_cookie') {
            foundCookie = true;
            console.log("Found _cookie:", value);
        }
    });
    
    if (!foundCookie) {
        console.log("_cookie not found in browser");
    }
    
    return foundCookie;
}

// Call this function after login to check if cookie was set
window.checkCookieInBrowser = checkCookieInBrowser;