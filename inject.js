// window.onload = function () {
	var req = new XMLHttpRequest();
	req.open("post", "http://192.168.1.200:8000/");
	req.setRequestHeader("Content-type", "application/x-www-form-urlencoded");
	req.send("action=enter&url=" + window.location + "&referrer=" + document.referrer);
	window.onbeforeunload = function () {
		req = new XMLHttpRequest();
		req.open("post", "http://192.168.1.200:8000", true);
		req.setRequestHeader("Content-type", "application/x-www-form-urlencoded");
		req.send("action=leave");
	}
// }
