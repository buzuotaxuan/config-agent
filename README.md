# 项目介绍

项目地址: https://github.com/lipg/config-agent

配置中心工具，目前支持通过http，https的方式从git中下载配置文件。

gitlab支持在线Web修改文件，遂使用该方式可实现一个支持传统应用的配置中心。使用shell
脚本也可是实现该功能，但是为了提供更好的兼容性，遂使用Go专门写的Agent，不依赖与系统环境组件，如：curl、wget等。

基本流程：`从Gitlab中获取访问Api的Access Key` -> `使用Config下载配置文件` -> `将配置文件复制并覆盖到指定位置` -> `启动应用`

# 项目结构

## git配置仓库结构

# 使用方式

## 配置文件下载脚本 `Get_Configuration.sh`

```bash
#!/usr/bin/env bash
# 配置文件配置
# 配置文件所在的服务器,可使用环境变量AUTOZI_CONFIG_URL注入。
export DEFAULT_AUTOZI_CONFIG_URL='http://git.lipg.cn'
# 配置文件仓库名称,可使用环境变量AUTOZI_CONFIG_REPO注入。
export DEFAULT_AUTOZI_CONFIG_REPO='cloud/config-agent'
# 配置文件标示，可以使用环境变量AUTOZI_CONFIG_TARGET注入。(可以是分支名称、Tag、可以是commit ID。)
export DEFAULT_AUTOZI_CONFIG_TARGET='master'
# 授权的Token,可以使用环境变量AUTOZI_CONFIG_TOKEN注入。
export DEFAULT_AUTOZI_CONFIG_TOKEN=''
# 配置文件下载的临时目录,可以使用环境变量CONFIG_TMP_PATH注入
export AUTOZI_CONFIG_TMP_PATH='/tmp'

# 检查配置参数，如果未注入，则使用默认值
if [ -z "$AUTOZI_CONFIG_URL" ]; then
    export AUTOZI_CONFIG_URL="$DEFAULT_AUTOZI_CONFIG_URL"
fi
if [ -z "$AUTOZI_CONFIG_REPO" ]; then
    export AUTOZI_CONFIG_REPO="$DEFAULT_AUTOZI_CONFIG_REPO"
fi
if [ -z "$AUTOZI_CONFIG_TARGET" ]; then
    export AUTOZI_CONFIG_TARGET="$DEFAULT_AUTOZI_CONFIG_TARGET"
fi
if [ -z "$AUTOZI_CONFIG_TOKEN" ]; then
    export AUTOZI_CONFIG_TOKEN="$DEFAULT_AUTOZI_CONFIG_TOKEN"
fi
if [ -z "$AUTOZI_CONFIG_TMP_PATH" ]; then
    export AUTOZI_CONFIG_TMP_PATH="$DEFAULT_AUTOZI_CONFIG_TMP_PATH"
fi
# 下载配置到临时目录
mkdir -pv ${AUTOZI_CONFIG_TMP_PATH}
$bash_dir/config --url=${AUTOZI_CONFIG_URL} --repo=${AUTOZI_CONFIG_REPO} --branch=${AUTOZI_CONFIG_TARGET} --token=${AUTOZI_CONFIG_TOKEN} --path=${AUTOZI_CONFIG_TMP_PATH}/config.tar
# 将配置文件展开
cd ${AUTOZI_CONFIG_TMP_PATH} && tar -xvf config.tar --strip-components=1 && rm -rf config.tar && ls -lR
```

## 更新配置文件 `App_Configuration.sh`

    将仓库中的app目录下的配置文件全部拷贝到指定目录，使用AUTOZI_CONFIG_PATH环境变量指定目录。

```bash
#!/usr/bin/env bash
# 检查目标环境
# 检查原始配置文件目录是否存在
if [ ! -d "$AUTOZI_CONFIG_TMP_PATH"  ] || [ ! -d "$AUTOZI_CONFIG_TMP_PATH/app"  ] ; then
    echo '配置文件不存在，将不处理任何配置';
fi
# 检查目标配置文件目录是否存在,不存在的话则创建目录
if [ ! -d "$AUTOZI_CONFIG_PATH" ]; then
    mkdir -pv ${AUTOZI_CONFIG_PATH};
fi
# 拷贝配置文件到指定目录
\cp -rv ${AUTOZI_CONFIG_TMP_PATH}/app/* ${AUTOZI_CONFIG_PATH}/
echo "完成APP配置文件处理。应用配置文件:$AUTOZI_CONFIG_TMP_PATH/app，目标路径:$AUTOZI_CONFIG_PATH";
```

## 扩展脚本

### 通用启动脚本 `entrypoint.sh`

```bash
#!/usr/bin/env bash
# 遇到任意命令返回非0则推出脚本
set -e
# 获取脚本所在的路径
bash_dir=$(cd `dirname $0`; pwd)
echo "启动脚本所在路径:$bash_dir";
# 设置默认参数
# 配置文件存储路径
export DEFAULT_AUTOZI_CONFIG_TMP_PATH='/tmp/config';
# 默认的配置文件路径,公共配置
export DEFAULT_AUTOZI_CONFIG_PATH='/usr/local/tomcat/webapps/ROOT/WEB-INF/classes';
# 处理公共配置文件
# 检查目标环境
if [ -z "$AUTOZI_CONFIG_TMP_PATH" ]; then
    export AUTOZI_CONFIG_TMP_PATH="$DEFAULT_AUTOZI_CONFIG_TMP_PATH";
fi
if [ -z "$AUTOZI_CONFIG_PATH" ]; then
    export AUTOZI_CONFIG_PATH="$DEFAULT_AUTOZI_CONFIG_PATH";
fi
echo "完成公共配置参数处理。AUTOZI_CONFIG_TMP_PACH:$AUTOZI_CONFIG_TMP_PATH ; AUTOZI_CONF_PATH:$AUTOZI_CONFIG_PATH";
# 更新全部配置文件
. ${bash_dir}/Get_Configuration.sh
# 处理App配置
. ${bash_dir}/App_Configuration.sh
# 处理Jar包配置文件
. ${bash_dir}/Jar_Reconfigure.sh
# 处理Tomcat配置
. ${bash_dir}/Tomcat_Configuration.sh
# 处理各环境独立脚本
. ${bash_dir}/Script_Configuration.sh

# 执行启动命令
exec "$@"
```

### 处理Jar包内的配置文件 `Jar_Reconfigure.sh`

```bash
#!/usr/bin/env bash
# 默认的lib依赖路径
export DEFAULT_APP_LIB_PATH='/usr/local/tomcat/webapps/ROOT/WEB-INF/lib';
# 检查原始配置文件目录是否存在
# 检查依赖资源目录
if [ -z "$APP_LIB_PATH" ]; then
    export APP_LIB_PATH="$DEFAULT_APP_LIB_PATH";
fi
Reconfigure(){
    # 查找Service包
    export JAR_FILE=$(find ${APP_LIB_PATH} -iregex ${JAR_TARGET_REGEX} | xargs ls -ltr | awk {'print $9'} | tail -1);
    echo "找到Service包，路径为:$JAR_FILE";
    # 跳转到配置文件目录
    cd ${AUTOZI_CONFIG_TMP_PATH}/${JAR_CONF_PATH};
    # 重新Service配置
    jar -uvf ${JAR_FILE} ./*;
    echo "完成Jar: $JAR_CONF_PATH 配置文件处理。应用配置文件:$AUTOZI_CONFIG_TMP_PATH/$JAR_CONF_PATH，目标路径:$AUTOZI_CLASS_PATH" 
}
# 处理Commons-Memcached的配置
export JAR_CONF_PATH='service-interface';
export JAR_TARGET_REGEX='.*/lzfinance-bt-service-interface-.*.jar';
Reconfigure;
# 处理Commons-Memcached的配置
export JAR_CONF_PATH='service-interface';
export JAR_TARGET_REGEX='.*/lzfinance-bt-service-interface-.*.jar';
Reconfigure;
.............
```

### 执行相关启动脚本，并打印所执行命令 `Script_Configuration.sh`

    需要执行的脚本可以通过注入环境变量AUTOZI_SCRIPT_FILE来启动，脚本名称使用都好分割，如AUTOZI_SCRIPT_FILE=link.sh,del.sh,mv.sh,nfs.sh

```bash
#!/usr/bin/env bash
# 默认执行的脚本
export DEFAULT_AUTOZI_SCRIPT_FILE=('links');
# 检查脚本是否存在
if [ ! -d "$AUTOZI_CONFIG_TMP_PATH"  ] || [ ! -d "$AUTOZI_CONFIG_TMP_PATH/script"  ] ; then
    echo '配置脚本不存在，将不处理任何脚本!'
fi
# 检查初始化脚本的环境变量
if [ -z "$AUTOZI_SCRIPT_FILE" ]; then
    AUTOZI_SCRIPT_FILE=${DEFAULT_AUTOZI_SCRIPT_FILE[@]}
else
    # 从环境变量获取需要执行的脚本
    OLD_IFS="$IFS" 
    IFS="," 
    AUTOZI_SCRIPT_FILE=(${AUTOZI_SCRIPT_FILE})
    IFS="$OLD_IFS" 
fi
# 再次确认脚本
if [ ! ${AUTOZI_SCRIPT_FILE} ]; then
    AUTOZI_SCRIPT_FILE=()
fi
set -x
# 载入软件环境变量配置
for script in ${AUTOZI_SCRIPT_FILE[@]};do
    if [ -f "$AUTOZI_CONFIG_TMP_PATH/script/$script.sh" ]; then
        source "$AUTOZI_CONFIG_TMP_PATH/script/$script.sh" 
    else
        echo "文件$AUTOZI_CONFIG_TMP_PATH/script/$script.sh不存在，不进行载入。" 
    fi
done
set +x
```

### 启动后阻塞进程退出 `sleep.sh`

```bash
#!/usr/bin/env bash
# 遇到任意命令返回非0则推出脚本
set -e
echo '完成文件处理,启动bash阻塞容器';
# 使用死循环，阻止容器退出
while true;
do
    sleep 24h;
done
```

### 启动时处理Tomcat端口 `Tomcat_Configuration.sh`

```bash
#!/usr/bin/env bash
# Tomcat端口配置
export DEFAULT_TOMCAT_HTTP_PORT='8080';
export DEFAULT_TOMCAT_AJP_PORT='8009';
export DEFAULT_TOMCAT_SERVER_PORT='8005';
# 检查目标环境
if [ -z "$TOMCAT_HTTP_PORT" ]; then
    export TOMCAT_HTTP_PORT="$DEFAULT_TOMCAT_HTTP_PORT";
fi
if [ -z "$TOMCAT_AJP_PORT" ]; then
    export TOMCAT_AJP_PORT="$DEFAULT_TOMCAT_AJP_PORT";
fi
if [ -z "$TOMCAT_SERVER_PORT" ]; then
    export TOMCAT_SERVER_PORT="$DEFAULT_TOMCAT_SERVER_PORT";
fi
# 打印并重写配置
sed -n "s/Connector port=\"8080\"/Connector port=\"$TOMCAT_HTTP_PORT\"/"p /usr/local/tomcat/conf/server.xml
sed -i "s/Connector port=\"8080\"/Connector port=\"$TOMCAT_HTTP_PORT\"/" /usr/local/tomcat/conf/server.xml
sed -n "s/Connector port=\"8009\"/Connector port=\"$TOMCAT_AJP_PORT\"/"p /usr/local/tomcat/conf/server.xml
sed -i "s/Connector port=\"8009\"/Connector port=\"$TOMCAT_AJP_PORT\"/" /usr/local/tomcat/conf/server.xml
sed -n "s/Server port=\"8005\"/Server port=\"$TOMCAT_SERVER_PORT\"/"p /usr/local/tomcat/conf/server.xml
sed -i "s/Server port=\"8005\"/Server port=\"$TOMCAT_SERVER_PORT\"/" /usr/local/tomcat/conf/server.xml
echo "即将启动Tomcat,并使用端口HTTP:$TOMCAT_HTTP_PORT,AJP:$TOMCAT_AJP_PORT,Server:$TOMCAT_SERVER_PORT";
```

## 启动时解压jar包指定文件到项目，优先加载从而解决加载顺序问题 `Jar_Uncompress.sh`

```bash
#!/usr/bin/env bash
# 默认的类文件文件路径
export DEFAULT_AUTOZI_CLASS_PATH='/usr/local/tomcat/webapps/ROOT/WEB-INF/classes';
# 默认的lib依赖路径
export DEFAULT_APP_LIB_PATH='/usr/local/tomcat/webapps/ROOT/WEB-INF/lib';
# 检查需要解压的资源目录
if [ -z "$APP_LIB_PATH" ]; then
    export APP_LIB_PATH="$DEFAULT_APP_LIB_PATH";
fi
# 检查解压所需要的文件
if [ -z "$AUTOZI_CLASS_PATH" ]; then
    export AUTOZI_CLASS_PATH="$DEFAULT_AUTOZI_CLASS_PATH";
fi
# 检查解压所需要的文件
if [ -z "$JAR_UNCOMPRESS_REGEX" ]; then
    export JAR_UNCOMPRESS_REGEX='./*';
fi
# 解压Jar文件
Uncompress(){
    # 查找Service包
    export JAR_FILE=$(find ${APP_LIB_PATH} -iregex ${SERVICE_TARGET_REGEX} | xargs ls -ltr | awk {'print $9'} | tail -1);
    echo "找到需要解压的Jar包，路径为:$JAR_FILE";
    # 解压相关Service
    # 创建临时目录
    mkdir -pv /tmp/uncompress && cd /tmp/uncompress;
    # 解压并拷贝文件
    jar -xvf ${JAR_FILE};
    \cp -rv ${JAR_UNCOMPRESS_REGEX} ${AUTOZI_CLASS_PATH};
    # 清理文件
    cd /tmp && rm -rf /tmp/uncompress;
    echo "完成jar: $JAR_FILE 解压处理。处理文件: $JAR_UNCOMPRESS_REGEX,目标路径:$AUTOZI_CLASS_PATH";
}
# 处理文件配置
export JAR_TARGET_REGEX='.*/autozi-common-core-.*.jar';
export JAR_UNCOMPRESS_REGEX='./org';
Uncompress;
# 处理文件配置
export JAR_TARGET_REGEX='.*/autozi-common-core-.*.jar';
export JAR_UNCOMPRESS_REGEX='./org';
Uncompress;
.........
```