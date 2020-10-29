<?php
/**
 * Copyright 2016 Google Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
/*
 * This is needed for cookie based authentication to encrypt password in
 * cookie
 * http://www.question-defense.com/tools/phpmyadmin-blowfish-secret-generator
 */
$cfg['blowfish_secret'] = '{{your_secret}}'; /* YOU MUST FILL IN THIS FOR COOKIE AUTH! */
/*
 * Servers configuration
 */
$i = 0;

// Change this to use the project and instance that you've created.
$host = '/cloudsql/livehub-277906:prod';
$type = 'socket';

/*
* First server
*/
$i++;
// IP address of your instance
$cfg['Servers'][$i]['host'] = '172.26.128.3';
// Use SSL for connection
$cfg['Servers'][$i]['ssl'] = true;
// Client secret key
$cfg['Servers'][$i]['ssl_key'] = '/etc/config/client-key.pem';
// Client certificate
$cfg['Servers'][$i]['ssl_cert'] = '/etc/config/client-cert.pem';
// Server certification authority
$cfg['Servers'][$i]['ssl_ca'] = '/etc/config/server-ca.pem';
// Disable SSL verification (see above note)
$cfg['Servers'][$i]['ssl_verify'] = false;

/*
 * End of servers configuration
 */

/*
 * Directories for saving/loading files from server
 */
$cfg['UploadDir'] = '';
$cfg['SaveDir'] = '';

/*
* Other settings
*/
$cfg['PmaNoRelation_DisableWarning'] = true;
$cfg['ExecTimeLimit'] = 60;
$cfg['CheckConfigurationPermissions'] = false;



