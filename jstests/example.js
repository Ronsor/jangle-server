/*Variable area*/
var Discord = require('./discordio.js');
var bot = new Discord.Client({
	token: process.env.TOKEN,
	autorun: true
});

/*Event area*/

bot.on("debug", console.log)

bot.on("ready", function(event) {
	console.log("Connected!");
	console.log("Logged in as: ");
	console.log(bot.username + " - (" + bot.id + ")");
	console.log(event.t);
});

bot.on("message", function(user, userID, channelID, message, event) {
	console.log(user + " - " + userID);
	console.log("in " + channelID);
	console.log(message);
	console.log("----------");

	if (message === "ping") {
		console.log("Got ping")
		sendMessages(channelID, ["Pong"]); //Sending a message with our helper function
	} else if (message === "picture") {
		sendFiles(channelID, ["fillsquare.png"]); //Sending a file with our helper function
	}
});

bot.on("presence", function(user, userID, status, game, event) {
	/*console.log(user + " is now: " + status);*/
});

bot.on("any", function(event) {
	//console.log(event); //Logs every event
});

bot.on("disconnect", function(x) {
	console.log("Bot disconnected", x);
	/*bot.connect()*/ //Auto reconnect
});

/*Function declaration area*/
function sendMessages(ID, messageArr, interval) {
	var resArr = [], len = messageArr.length;
	var callback = typeof(arguments[2]) === 'function' ?  arguments[2] :  arguments[3];
	if (typeof(interval) !== 'number') interval = 1000;

	function _sendMessages() {
		setTimeout(function() {
			if (messageArr[0]) {
				bot.sendMessage({
					to: ID,
					message: messageArr.shift()
				}, function(err, res) {
					resArr.push(err || res);
					if (resArr.length === len) if (typeof(callback) === 'function') callback(resArr);
				});
				_sendMessages();
			}
		}, interval);
	}
	_sendMessages();
}

function sendFiles(channelID, fileArr, interval) {
	var resArr = [], len = fileArr.length;
	var callback = typeof(arguments[2]) === 'function' ? arguments[2] : arguments[3];
	if (typeof(interval) !== 'number') interval = 1000;

	function _sendFiles() {
		setTimeout(function() {
			if (fileArr[0]) {
				bot.uploadFile({
					to: channelID,
					file: fileArr.shift()
				}, function(err, res) {
					resArr.push(err || res);
					if (resArr.length === len) if (typeof(callback) === 'function') callback(resArr);
				});
				_sendFiles();
			}
		}, interval);
	}
	_sendFiles();
}
