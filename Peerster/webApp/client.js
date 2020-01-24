$(document).ready(function(){

	var guiport = prompt("PLEASE INPUT THE GUIPORT", 13081)
	var url = "http://127.0.0.1:" + guiport + "/"
	// Get id
	$.ajax({
		url: url + "id",
		type: "GET",
		dataType: "json",
		success: function(json) {

			console.log(json)
			var ID = json.id
			console.log("id is " + ID)
			$("#NodeID").append("<P>" + ID)
		}
	})

	// Initialize buffer for msg and peernodes
	var peer_addrs = new Array();
	var rumors = new Array();
	var routable = new Array();
	var searched = new Array();

	// Define update peer_addrs and rumors func
	function updatePeers() {

		$.ajax({
			url: url + "node",
			type: "GET",
			dataType: "json",
			success: function(json){

				console.log(json)
				var new_peer_addrs = Array.from(json.nodes, x => x);

				// Check for updates
				if (new_peer_addrs.length > peer_addrs.length) {

					peer_addrs = new_peer_addrs;

					// Clear old display
					$("#PeerBox").empty();

					// Create new display
					peer_addrs.forEach((v, i) => {

						$("#PeerBox").append("<P>" + v);
					});
				}

 			}
 		});
	};

	function updateRoutable() {

		$.ajax({
			url: url + "routing",
			type: "GET",
			dataType: "json",
			success: function(json){

				console.log(json);
				var new_routable_peers = Array.from(json.nodes, x => x);

				if (new_routable_peers.length > routable.length) {

					routable = new_routable_peers;

					$("#RoutableBox").empty();

					routable.forEach((v, i) => {

						console.log("Adding routable")
						$("#RoutableBox").append($("<option/>", {
							value : v,
							text : v
						}));
					})
				}
			} 
		})
	}
	function updateMsg() {

		$.ajax({
			url: url + "message",
			type: "GET",
			dataType: "json",
			success: function(json){

				var new_updated_rumors = Array.from(json.messages, x => x);
				console.log(new_updated_rumors)
				// Check for updates
				if (new_updated_rumors.length > rumors.length) {

					rumors = new_updated_rumors;
					
					$("#MsgBox").empty();

					rumors.forEach((v, i) => {
						console.log(v)
						$("#MsgBox").append("<p class='message'>" + v + "</p>");
					});
				}
 			}
 		});
	}
	function updateSearch() {

		$.ajax({
			url: url + "search",
			type: "GET",
			dataType: "json",
			success: function(json){
				var new_updated_matched = Array.from(json.matches, x => x);
				console.log(new_updated_matched)
				// Check for updates
				if (new_updated_matched.length > searched.length) {

					searched = new_updated_matched;

					$("#MatchedBox").empty();

					searched.forEach((v, i) => {

						$("#MatchedBox").append($("<option/>", {
							value : v.substring(13,),
							text : v.substring(13,)
						}));
					});
				}
 			}
 		});
	}
	// Run update periodically
	setInterval(updatePeers, 1000);
	setInterval(updateMsg, 1000);
	setInterval(updateRoutable, 1000);
	setInterval(updateSearch, 1000)
	// Define handler for add msg
	$("#InputBtn").click(function(){

		// Get the text in textarea
		var text = $("#InputMsg").val();

		// Refresh textarea
		$("#InputMsg").val("");

		// Send text to server
		var data = {text : text};

		$.ajax({
			url: url + "message",
			type: "POST",
			data: JSON.stringify(data),
			dataType: "json",
			success: function(msg) {

				alert("Successfully add msg!!!")
			}
		});
	});

	// Define handler for add private msg
	$("#PrivateBtn").click(function(){

		// Get destination and text
		var dest = $("#RoutableBox").children("option:selected").val();
		var text = $("#PrivateInput").val();

		// Refresh text area
		$("#PrivateInput").val("");

		// Send msg to server
		var data = {Text : text,
					Dest : dest};

		$.ajax({

			url: url + "routing",
			type: "POST",
			data: JSON.stringify(data),
			dataType: "json",
			success: function(msg) {

				alert("Sucessfully send private msg");
			}
		});
	})

	// Define handler for file sharing
	$("#ShareBtn").click(function(){

		// Get destination of file
		var dest = $("#ToShareFile").val();

		// Refresh to share file
		$("#ToShareFile").val("");

		// Send fileName to fileSharer
		var data = {
			Name : dest
		};

		$.ajax({

			url: url + "sharing",
			type: "POST",
			data: JSON.stringify(data),
			dataType: "json",
			success: function(msg) {

				alert("Sucessfully share file")
			},
			error: function(xhr, status) {

				alert("Input fileName not valid!!!")
			}
		})
	})
	// Define handler for add pere
	$("#PeerAddBtn").click(function(){

		// Get the text in the textarea
		var text = $("#PeerAddr").val();

		// Refresh peer addr
		$("#PeerAddr").val("");

		var regExp = RegExp("((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)");
		console.log("New peer adding feature!!");
		console.log(regExp.test(text));
		if (!regExp.test(text)) {

			alert("Bad IP address!!!");
			return;
		}
		// Send new addr to server
		var data = {addr : text};

		$.ajax({
			url: url + "node",
			type: "POST",
			data: JSON.stringify(data),
			dataType: "json",
			success: function(msg) {

				alert("Successfully add peer");
			}
		});
	});

	// Define handler for file request
	$("#RequestBtn").click(function(){

		// Get target peer name, local file name and metahash
		var target = $("#RequestPeer").val();
		var localFileName = $("#FileNameToStore").val();
		var metaHash = $("#MetaHash").val();

		// Fresh the input
		$("#RequestPeer").val("")
		$("#FileNameToStore").val("")
		$("#MetaHash").val("")

		// Send request to peer
		var data = {

			Dest : target,
			FileName : localFileName,
			MetaHash : metaHash
		};

		$.ajax({

			url: url + "request",
			type: "POST",
			data: JSON.stringify(data),	
			dataType: "json",
			success: function(msg) {

				alert("Successfully request a file");
			}
		});
	})

	// Define handler for file search
	$("#SearchBtn").click(function() {

		// Get keywords to search
		var keywords = $("#SearchFile").val();
		$("#SearchFile").val("");

		// Send request to peer
		var data = {
			Keywords : keywords,
		};
		$.ajax({
			url: url + "search",
			type: "POST",
			data: JSON.stringify(data),
			dataType: "json",
			success: function(msg) {
				alert("Successfully search for matches of keywords");
			}
		})
	})

	// Define handler for downloading searched file
	$("#DownloadSearched").click(function(){
		var target = $("#MatchedBox").children("option:selected").val();
		var data = {
			Name: target,
		};
		console.log("Downloading" + target);
		$.ajax({
			url: url + "download",
			type: "POST",
			data: JSON.stringify(data),
			dataType: "json",
			success: function(msg) {
				alert("Successfully Download the file");
			}
		});
	})
})

