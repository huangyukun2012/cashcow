<?php
//please change the following link tu you country!!!!!!!
//please change the following link tu you country!!!!!!!
//please change the following link tu you country!!!!!!!
//important words need to say 3 times
define('GOOGLE_URL', 'https://translate.google.cn/translate_a/');


define('UA', isset($_SERVER['HTTP_USER_AGENT']) && !empty($_SERVER['HTTP_USER_AGENT']) ?  $_SERVER['HTTP_USER_AGENT'] : 'Mozilla/5.0 (Android; Mobile; rv:22.0) Gecko/22.0 Firefox/22.0');


function shr32($x, $bits)
{

    if($bits <= 0){
        return $x;
    }
    if($bits >= 32){
        return 0;
    }

    $bin = decbin($x);
    $l = strlen($bin);

    if($l > 32){
        $bin = substr($bin, $l - 32, 32);
    }elseif($l < 32){
        $bin = str_pad($bin, 32, '0', STR_PAD_LEFT);
    }

    return bindec(str_pad(substr($bin, 0, 32 - $bits), 32, '0', STR_PAD_LEFT));
}        

function charCodeAt($str, $index)
{
    $char = mb_substr($str, $index, 1, 'UTF-8');
 
    if (mb_check_encoding($char, 'UTF-8'))
    {
        $ret = mb_convert_encoding($char, 'UTF-32BE', 'UTF-8');
        return hexdec(bin2hex($ret));
    }
    else
    {
        return null;
    }
}

function mb_str_split($str, $length = 1) 
{
    if ($length < 1) return false;
    $result = array();
    for ($i = 0; $i < mb_strlen($str); $i += $length) {
        $result[] = mb_substr($str, $i, $length);
    }
    return $result;
}
        
function RL($a, $b)
{
    for($c = 0; $c < strlen($b) - 2; $c +=3) {
        $d = $b{$c+2};
        $d = $d >= 'a' ? charCodeAt($d,0) - 87 : intval($d);
        $d = $b{$c+1} == '+' ? shr32($a, $d) : $a << $d;
        $a = $b{$c} == '+' ? ($a + $d & 4294967295) : $a ^ $d;
    }
    return $a;
}


function sendHttpRequest($url, $post, $requestBody, $headers = null) 
{
    $ch = curl_init();
    curl_setopt($ch, CURLOPT_VERBOSE, 1);
    curl_setopt($ch, CURLOPT_URL, $url);
    curl_setopt($ch, CURLOPT_SSL_VERIFYPEER, 0);
    curl_setopt($ch, CURLOPT_SSL_VERIFYHOST, 0);

    if ( $headers ) curl_setopt($ch, CURLOPT_HTTPHEADER, $headers);

    if ($post) {
        curl_setopt($ch, CURLOPT_POST, 1);
        curl_setopt($ch, CURLOPT_POSTFIELDS, $requestBody);
    }

    curl_setopt($ch, CURLOPT_RETURNTRANSFER, 1);
    curl_setopt($ch, CURLOPT_TIMEOUT, 30);

    $response = curl_exec($ch);
    $curl_errno = curl_errno($ch); 
    $curl_error = curl_error($ch); 
    curl_close($ch);
    
    if( $curl_errno > 0 ){ 
        return "$curl_error";
    }       

    return $response;
}

function TKK() 
{
    $a = 561666268;
    $b = 1526272306;
    return 406398 . '.' . ($a + $b);
}

function TL($a) 
{
    
    $tkk = explode('.', TKK());
    $b = $tkk[0];

    for($d = array(), $e = 0, $f = 0; $f < mb_strlen ( $a, 'UTF-8' ); $f ++) {
        $g = charCodeAt ( $a, $f );
        if (128 > $g) {
            $d [$e ++] = $g;
        } else {
            if (2048 > $g) {
                $d [$e ++] = $g >> 6 | 192;
            } else {
                if (55296 == ($g & 64512) && $f + 1 < mb_strlen ( $a, 'UTF-8' ) && 56320 == (charCodeAt ( $a, $f + 1 ) & 64512)) {
                    $g = 65536 + (($g & 1023) << 10) + (charCodeAt ( $a, ++ $f ) & 1023);
                    $d [$e ++] = $g >> 18 | 240;
                    $d [$e ++] = $g >> 12 & 63 | 128;
                } else {
                    $d [$e ++] = $g >> 12 | 224;
                    $d [$e ++] = $g >> 6 & 63 | 128;
                }
            }
            $d [$e ++] = $g & 63 | 128;
        }
    }
    $a = $b;
    for($e = 0; $e < count ( $d ); $e ++) {
        $a += $d [$e];
        $a = RL ( $a, '+-a^+6' );
    }
    $a = RL ( $a, "+-3^+b+-f" );
    $a ^= $tkk[1];
    if (0 > $a) {
        $a = ($a & 2147483647) + 2147483648;
    }
    $a = fmod ( $a, pow ( 10, 6 ) );
    return $a . "." . ($a ^ $b);
}

function translate($sl, $tl, $q, $param = 't?client=webapp', $method = 'get')
{
    
    $tk = TL($q);
    $q = urlencode(stripslashes($q));
    $resultRegexes = array(
        '/,+/'  => ',',
        '/\[,/' => '[',
    );    
    
    $url = GOOGLE_URL . $param . "&sl=".$sl."&tl=".$tl."&hl=".$tl."&dt=bd&dt=ex&dt=ld&dt=md&dt=qca&dt=rw&dt=rm&dt=ss&dt=t&dt=at&ie=UTF-8&oe=UTF-8&otf=2&ssel=0&tsel=0&kc=1&tk=". $tk ;
    if ( $method == 'get' ) $url .= "&q=" . $q;
    $output = sendHttpRequest($url, $method == 'get' ? 0 : 1, $method == 'get' ? '' : "&q=" . $q, array('User-Agent' => UA));
    
    return $output;
}

$q = file_get_contents('./input', true);
file_put_contents("output", "");

$sl = 'en';
$tl = 'zh-CN';
$p =  '1';
$method = 'post';

$type = 'output';


if ( $p == 1 ) {
    $param = 't?client=webapp';
}
if ( $p == 2 ) {
    $param = 'single?client=t';
}
if ( isset($_GET['txt']) ) {
    $q = $_GET['txt'];
    $type = 'json';
}

$translate = translate($sl, $tl, $q, $param, $method);

if ( $type == 'json' ) {
    if ( isset($_GET['txt']) ) {
        echo json_encode(array('ret' => $translate));
    } else {
        echo json_encode(array('tk' => TL($q), 'translate' => $translate));
    }
    exit;
}

echo file_put_contents("output", $translate);

?>