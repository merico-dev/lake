# Gitee Pond

<div align="center">

| [English](README.md) | [中文](README-zh-CN.md) |
| --- | --- |

</div>

<br>

## Summary

## Configuration

### Provider (Datasource) Connection
The connection aspect of the configuration screen requires the following key fields to connect to the **Gitee API**. As gitee is a _single-source data provider_ at the moment, the connection name is read-only as there is only one instance to manage. As we continue our development roadmap we may enable _multi-source_ connections for gitee in the future.

- **Connection Name** [`READONLY`]
    - ⚠️ Defaults to "**Gitee**" and may not be changed.
- **Endpoint URL** (REST URL, starts with `https://` or `http://`)
    - This should be a valid REST API Endpoint eg. `https://gitee.com/api/v5/`
    - ⚠️ URL should end with`/`
- **Auth Token(s)** (Personal Access Token)
    - For help on **Creating a personal access token**
    - Provide at least one token for Authentication with the . This field accepts a comma-separated list of values for multiple tokens. The data collection will take longer for gitee since they have a **rate limit of 2k requests per hour**. You can accelerate the process by configuring _multiple_ personal access tokens.

"For API requests using `Basic Authentication` or `OAuth`


If you have a need for more api rate limits, you can set many tokens in the config file and we will use all of your tokens.

For an overview of the **gitee REST API**, please see official [gitee Docs on REST](https://gitee.com/api/v5/swagger)

Click **Save Connection** to update connection settings.


### Provider (Datasource) Settings
Manage additional settings and options for the gitee Datasource Provider. Currently there is only one **optional** setting, *Proxy URL*. If you are behind a corporate firewall or VPN you may need to utilize a proxy server.

**gitee Proxy URL [ `Optional`]**
Enter a valid proxy server address on your Network, e.g. `http://your-proxy-server.com:1080`

Click **Save Settings** to update additional settings.

### Regular Expression Configuration
Define regex pattern in .env
- GITEE_PR_BODY_CLOSE_PATTERN: Define key word to associate issue in pr body, please check the example in .env.example

## Sample Request
In order to collect data, you have to compose a JSON looks like following one, and send it by selecting `Advanced Mode` on `Create Pipeline Run` page:
1. Configure-UI Mode
```json
[
  [
    {
      "plugin": "gitee",
      "options": {
        "repo": "lake",
        "owner": "merico-dev",
        "token": "xxxx"
      }
    }
  ]
]
```
and if you want to perform certain subtasks.
```json
[
  [
    {
      "plugin": "gitee",
      "subtasks": ["collectXXX", "extractXXX", "convertXXX"],
      "options": {
        "repo": "lake",
        "owner": "merico-dev",
        "token": "xxxx"
      }
    }
  ]
]
```

2. Curl Mode:
   You can also trigger data collection by making a POST request to `/pipelines`.
```
curl --location --request POST 'localhost:8080/pipelines' \
--header 'Content-Type: application/json' \
--data-raw '
{
    "name": "gitee 20211126",
    "tasks": [[{
        "plugin": "gitee",
        "options": {
            "repo": "lake",
            "owner": "merico-dev"
            "token": "xxxx"
        }
    }]]
}
'
```
and if you want to perform certain subtasks.
```
curl --location --request POST 'localhost:8080/pipelines' \
--header 'Content-Type: application/json' \
--data-raw '
{
    "name": "gitee 20211126",
    "tasks": [[{
        "plugin": "gitee",
        "subtasks": ["collectXXX", "extractXXX", "convertXXX"],
        "options": {
            "repo": "lake",
            "owner": "merico-dev"
            "token": "xxxx"
        }
    }]]
}
'
```
