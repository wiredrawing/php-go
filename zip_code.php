<?php



$f = file("43KUMAMO.csv");

$sql_list = [];
$duplication_check = [];
foreach ($f as $key => $value) {
    $value = explode(",", $value);

    $value = array_map(function ($v) {
        return str_replace("\"", "", $v);
    }, $value);
    $zip_code = $value[2];
    $prefecture = $value[6];
    $city = $value[7];
    $town = $value[8];
    $block = str_replace("熊本市", "", $city);

    // 郵便番号の重複チェック
    if (in_array($zip_code, $duplication_check)) {
        continue;
    }

    $sql_list[] = "
    insert into zip_master
    (
        zip_code,
        prefecture,
        city,
        town,
        block
    )
    values (
        '{$zip_code}',
        '{$prefecture}',
        '{$city}',
        '{$town}',
        '{$block}'
    );
    ";
    $sql_list[] = PHP_EOL;
}
print_r($sql_list);
print(join("", $sql_list));
