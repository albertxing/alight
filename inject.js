(function () {var preq = new XMLHttpRequest();
preq.open("post", "http://192.168.1.200:8000/");
preq.setRequestHeader("Content-type", "application/x-www-form-urlencoded");
preq.onload = function () {
	if (preq.responseText) {
		var avid = JSON.parse(preq.responseText).vid;
		localStorage["avid"] = avid;
	}
}

var params = "action=enter&url=" + window.location + "&referrer=" + document.referrer;
if (localStorage && localStorage["avid"]) params += "&avid=" + localStorage["avid"];
preq.send(params);
window.onbeforeunload = function () {
	preq = new XMLHttpRequest();
	preq.open("post", "http://192.168.1.200:8000", true);
	preq.setRequestHeader("Content-type", "application/x-www-form-urlencoded");
	preq.send("action=leave");
})();
