/**
 * Copyright 2017 IBM All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 */
'use strict';
var userFile = require('../artifacts/users.json');
var helper = require('../app/helper.js');
var invoke = require('../app/invoke-transaction.js');
var jwt = require('jsonwebtoken');
var logger = helper.getLogger('user-setup');
var hfc = require('fabric-client');
var fs = require('fs');

var userRegistration = async function () {
    let users = userFile["users"];
    let userTokenObj = [];
    for (let index in users) {
        let userObj = users[index];
        var token = jwt.sign({
            exp: Math.floor(Date.now() / 1000) + parseInt(hfc.getConfigSetting('jwt_expiretime')),
            username: userObj["userId"],
            orgName: userObj["org"]
        }, "thisismysecret");
        //logger.debug('UserId %s : Token %s', userObj["userId"], token);
        let newUserObj = Object.assign({}, userObj);
        delete newUserObj["balance"];
        userTokenObj.push({ "user": newUserObj, "tokenId": token });
        let attributes = helper.convertUserAttributes(userObj, ["email", "org"]);
        let response = await helper.getRegisteredUser(userObj["userId"], userObj["org"], attributes, true);
        if (response && typeof response !== 'string') {
            logger.debug('Successfully registered the username %s for organization %s', userObj["userId"], userObj["org"]);
        } else {
            logger.debug('Failed to register the username %s for organization %s with::%s', userObj["userId"], userObj["org"], response);
        }
        let args = [JSON.stringify(userObj)];
        let message = await invoke.invokeChaincode(hfc.getConfigSetting('peers'), hfc.getConfigSetting('channelName'), hfc.getConfigSetting('chaincodeName'), "addUser", args, userObj["userId"], userObj["org"]);
        logger.debug('Response from invoke of addUser %s', message);
    }
    writeTokenToFile(userTokenObj);
};
var writeTokenToFile = function (userTokens) {
    var filePath = hfc.getConfigSetting('tokenFilePath');
    if (fs.existsSync(filePath)) {
        logger.debug("Removing existing token file");
        fs.unlinkSync(filePath);
    }
    var json = JSON.stringify(userTokens);
    var writeStream = fs.createWriteStream(filePath);
    writeStream.write(json);
    writeStream.end();
    logger.debug("Written tokens to file");
}

exports.userRegistration = userRegistration;
