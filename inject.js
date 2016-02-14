(function () {
	var url = "http://CHANGEME";
	var preq = new XMLHttpRequest();
	preq.open("post", url);
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
		preq.open("post", url, true);
		preq.setRequestHeader("Content-type", "application/x-www-form-urlencoded");
		preq.send("action=leave");
	}

})();
