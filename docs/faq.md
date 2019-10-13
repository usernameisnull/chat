# Frequently Asked Questions

### Q: Where can I find server logs when running in Docker?<br/>
**A**: The log is in the container at `/var/log/tinode.log`. Attach to a running container with command
```
docker exec -it name-of-the-running-container /bin/bash
```
Then, for instance, see the log with `tail -50 /var/log/tinode.log`

If the container has stopped already, you can copy the log out of the container (saving it to `./tinode.log`):
```
docker cp name-of-the-container:/var/log/tinode.log ./tinode.log
```

Alternatively, you can instruct the docker container to save the logs to a directory on the host by mapping a host directory to `/var/log/` in the container. Add `-v /where/to/save/logs:/var/log` to the `docker run` command.

### Q: How to setup FCM push notifications?<br/>
**A**: If you running the server directly:
1. Create a project at https://firebase.google.com/ if you have not done so already.
2. Follow instructions at https://cloud.google.com/iam/docs/creating-managing-service-account-keys to download the credentials file.
3. Update the server config [`tinode.conf`](../server/tinode.conf#L255), section `"push"` -> `"name": "fcm"`. Do _ONE_ of the following:
  * _Either_ enter the path to the downloaded credentials file into `"credentials_file"`.
  * _OR_ copy the file contents to `"credentials"`.<br/><br/>
    Remove the other entry. I.e. if you have updated `"credentials_file"`, remove `"credentials"` and vice versa.
4. Update [TinodeWeb](/tinode/webapp/) config [`firebase-init.js`](https://github.com/tinode/webapp/blob/master/firebase-init.js): update `messagingSenderId` and `messagingVapidKey`. See more info at https://github.com/tinode/webapp/#push_notifications
5. If you are using an Android client, add `google-services.json` to [Tindroid](/tinode/tindroid/) by following instructions at https://developers.google.com/android/guides/google-services-plugin and recompile the client. See more info at https://github.com/tinode/tindroid/#push_notifications
6. If you are using an iOS client, add `GoogleService-Info.plist` to [Tinodios](/tinode/ios/) by following instructions at https://firebase.google.com/docs/cloud-messaging/ios/client) and recompile the client. See more info at https://github.com/tinode/ios/#push_notifications

If you are using the [Docker image](https://hub.docker.com/u/tinode):
1. Create a project at https://firebase.google.com/ if you have not done so already.
2. Follow instructions at https://cloud.google.com/iam/docs/creating-managing-service-account-keys to download the credentials file.
3. Follow instructions in the Docker [README](../docker#enable-push-notifications) to enable push notifications for the `TinodeWeb`.
4. Add `google-services.json` to [Tindroid](/tinode/tindroid/#push_notifications), `GoogleService-Info.plist` to [Tinodios](/tinode/ios/#push_notifications), recompile the apps.


### Q: How can new users be added to Tinode?<br/>
**A**: There are three ways to create accounts:
* A user can create a new account using one of the applications (web, Android, iOS).
* A new account can be created using [tn-cli](../tn-cli/) (`acc` command). The process can be scripted.
* If the user already exists in an external database, the Tinode account can be automatically created on the first login using the [rest authenticator](../server/auth/rest/).


### Q: How to create a `root` user?<br/>
**A**: The `root` access can be granted to a user only by executing a database query. First create or choose the user you want to promote to `root` then execute the query:
* RethinkDB:
```js
r.db("tinode").table("auth").get("basic:login-of-the-user-to-make-root").update({authLvl: 30})
```
* MySQL:
```sql
UPDATE auth SET authlvl=30 WHERE uname='basic:login-of-the-user-to-make-root';
```
The test database has a stock user `xena` which has root access.
