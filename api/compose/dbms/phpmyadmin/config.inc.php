<?php
$i = 0;

// MySQL container
$i++;
$cfg['Servers'][$i]['host'] = 'mysql';
$cfg['Servers'][$i]['auth_type'] = 'cookie';
$cfg['Servers'][$i]['verbose'] = 'MySQL';

// MariaDB container
$i++;
$cfg['Servers'][$i]['host'] = 'mariadb';
$cfg['Servers'][$i]['auth_type'] = 'cookie';
$cfg['Servers'][$i]['verbose'] = 'MariaDB';
