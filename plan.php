<?php
ini_set("memory_limit", -1);
$question_sql_list = [];
$answer_sql_list = [];
foreach (range(1,10000) as $key => $value) {

    $t = date("Y-m-d H:i:s");
    $person_id = sprintf("X000000%d", $value);
    $sql = <<< EOF
        insert into questions (
            id,
            question,
            person_id,
            answered_at
        ) values (
            {$value},
            '質問例 : {$value}',
            '{$person_id}',
            '{$t}'
        );
EOF;
        $question_sql_list[] = $sql;
    foreach(range(1, 100)as $inner_key => $inner_value) {
        $answer_sql = <<< EOF
        insert into answers (
            id,
            question_id,
	    person_id,
	    answer,
            answered_at
        ) values (
            {$inner_value},
		{$value},
		'{$person_id}',
            '回答回答回答回答回答回答回答 : 回答しました{$value}',
            '{$t}'
        );
EOF;
        $answer_sql_list []  = $answer_sql;

    }
}


file_put_contents("question.sql", join(PHP_EOL, $question_sql_list));
file_put_contents("answer.sql", join(PHP_EOL, $answer_sql_list));
