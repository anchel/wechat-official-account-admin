

# **WeChat Official Account Management System**

## Background

WeChat Official Account provides various APIs to manage account resources, such as media management, user management, and QR codes. It also supports configuring a "Server URL" to take over message push notifications from the official account. This project implements these features.

Demo URL: [http://wechat.anchel.cn](http://wechat.anchel.cn/)

Username/Password: guest/guest

Screenshot:
![173130945135422](https://www.anchel.cn/images/wechat/upload-6735ef232b22bae98b20cfd7.jpeg)

## **Implemented Features**

1. Multi-account management
2. Message receiving (plain text, encrypted)
3. Message replying (text, image, audio, video, articles (external link))
4. Custom menus (including personalized menus)
5. Media management (images, audio, video)
6. QR codes (permanent, temporary)
7. User management (add tags, delete tags, set user tags, set user remarks)

### Note

The official WeChat Official Account management backend (mp.weixin.qq.com) offers comprehensive features and a polished user experience. In general, it is not recommended to deploy your own management system, as it may lack some features and user experience compared to the official platform.

## Technical Architecture

The project follows a frontend-backend separation architecture and is divided into two parts:

- Backend: Golang + MongoDB + Redis (this repository)
- Frontend: Vue3 + Element Plus, repository URL: https://github.com/anchel/wechat-official-account-admin-fe

The frontend project is embedded in the backend as a Git submodule. During the build process, the backend packages all frontend resources together, resulting in a single executable file.

## Deployment

There are two deployment methods: using a pre-built executable or building from source manually.

1. Using a pre-built executable

   a. Visit the [Download Page](https://github.com/anchel/wechat-official-account-admin/releases) and download the executable corresponding to your server type. For example, if your server is Linux with an AMD64 CPU, download woaa-linux-amd64. For convenience, rename the file to woaa

   b. Place the woaa executable in a directory on your server, e.g., /data/app

   c. In the /data/app directory, run the command: `chmod +x woaa` to grant executable permissions.

   d. Rename the `.env.example` file from the project source to `.env` and place it in the /data/app directory (same directory as woaa). Fill in your Redis and MongoDB connection details accordingly.

   e. In the /data/app directory, run `./woaa` to start the service.

2. Building from source manually

   a. Prerequisites: Git, Golang (1.22+), Node.js (20+)

   b. Clone the repository locally: `git clone --recurse-submodules https://github.com/anchel/wechat-official-account-admin.git`. The submodule will be cloned automatically.

   c. In the project root directory, run `make fe` to install frontend dependencies and build the frontend assets.

   d. Run `make all` to build executables for multiple platforms. After completion, the build directory will contain executables for various platforms. Choose the appropriate one for your environment.

   e. Follow the subsequent steps outlined in method 1 above.

## Usage

After starting the service, it listens on port 9305 by default. Assuming the IP is 192.168.0.100:

Open a browser and visit: http://192.168.0.100:9305

Upon the first launch, the system will automatically create a superadmin user `admin` with the password `admin1987`. You can log in with this account to create other users. Please change the superadmin password as soon as possible.

After logging in, there are no official accounts by default. You need to add one first. Follow these steps:

1. Go to the official WeChat Official Account backend [Official Backend](https://mp.weixin.qq.com/) to obtain your AppID, AppSecret, Token, and EncodingAESKey.
2. On the Official Account Management page of this system, add your account.

   ![1731309451354](https://www.anchel.cn/images/wechat/upload-6735ba422b22bae98b20cfd5.png)

3. After adding, view the details page to see the configuration URL. This URL must be configured in the official WeChat backend.

   ![1731309337778](https://www.anchel.cn/images/wechat/upload-6735ba622b22bae98b20cfd6.png)

   Additionally, WeChat restricts requests based on source IP. You must add the public IP displayed on the page to the IP whitelist in the official WeChat backend.

4. Once configured, you can start managing your official account.

## Troubleshooting

1. Some APIs indicate insufficient permissions:
   Official accounts are categorized into Subscription Accounts and Service Accounts, and are further differentiated by verification status. Different account types have different API permission scopes. For details, see [API Permission Guidelines](https://developers.weixin.qq.com/doc/offiaccount/Getting_Started/Explanation_of_interface_privileges.html)
2. Requester IP is not in the whitelist:
   You need to add your server's public IP to the whitelist in the official WeChat backend.
3. API frequency rate limited:
   WeChat imposes rate limits on API calls. For details, see [Daily Call Limits](https://mp.weixin.qq.com/cgi-bin/frame?t=pages/developsetting/page/developsetting_frame)
