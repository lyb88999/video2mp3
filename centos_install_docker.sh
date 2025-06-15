#!/bin/bash  
  
echo "欢迎使用centos一键安装docker脚本，本脚本在centos7上做过测试，其余版本不保证可用！"
# 提示用户输入  
echo "你想继续运行这个脚本吗？(y/n):"  
read user_input  
  
# 检查用户输入  
if [ "$user_input" = "y" ] || [ "$user_input" = "Y" ]; then  
    echo "用户选择继续运行脚本..."  
    echo "(1/4)开始安装docker所需包"
    sleep 3
    yum install -y yum-utils device-mapper-persistent-data lvm2
    echo "(2/4)完毕 开始切换国内阿里源"
    sleep 2
    yum-config-manager --add-repo http://mirrors.aliyun.com/docker-ce/linux/centos/docker-ce.repo
    echo "(3/4)完毕 开始安装docker服务，可能需要3分钟左右"
    sleep 2
    yum install docker-ce docker-ce-cli containerd.io
    echo "(4/4)完毕 设置docker开机自启"
    sleep 2
    systemctl enable docker
    echo "完毕 准备添加docker源 如果放弃请按crtl+c结束运行 docker源需要自行准备！"
    sleep 5
    vi /etc/docker/daemon.json
    echo "完毕 准备重载docker配置"
    sleep 2
    systemctl daemon-reload
    echo "完毕 开始重启docker"
    sleep 2
    systemctl restart docker
    echo "完毕 脚本运行完毕"
else  
    echo "用户选择退出脚本。"   
fi  
  
# 脚本的其他部分...  
# 如果用户选择继续，则这部分会被执行  
echo "脚本运行完毕。"
