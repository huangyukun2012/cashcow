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
