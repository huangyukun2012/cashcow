function is_mobile(){
	if (!navigator.userAgent.match(/(iPhone|iPod|Android|ios|iPad)/i)){
		return false;
	}else{
		return true;
	}
}

var params = {
	"sugSubmit": false
};
BaiduSuggestion.bind("search", params);



function search() {
	var si = document.getElementById("search").value;
	if (si.trim() != "")
	{
		window.location.href="/search/" + si;
	}
}

function keysearch(){
	if (event.keyCode==13)  //回车键的键值为13
		document.getElementById("searchbutton").click(); //调用登录按钮的登录事件
}
