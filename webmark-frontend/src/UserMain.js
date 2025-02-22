import React, { useState, useEffect } from 'react';
import { Layout, Input, Button, Flex, Space, Modal, message } from 'antd';
import {
    AppstoreAddOutlined,
    ExportOutlined,
    SearchOutlined,
    UserAddOutlined,
    SettingOutlined,
    LogoutOutlined
} from '@ant-design/icons';
import './UserMain.css';
import { useNavigate } from 'react-router-dom';
import Cookies from 'js-cookie';

const { Header, Content } = Layout;

const UserMain = () => {
    const navigate = useNavigate();
    const [messageApi, contextHolder] = message.useMessage();
    const [keywords, setKeywords] = useState("");
    const [groupList, setGroupList] = useState([]);
    // 新建分组模态框start
    const [newGroupData, setNewGroupData] = useState({});
    const [isNewGroupModalOpen, setIsNewGroupModalOpen] = useState(false);
    // 新建分组模态框end

    // 新建用户模态框start
    const [newUserData, setNewUserData] = useState({});
    const [isNewUserModalOpen, setIsNewUserModalOpen] = useState(false);
    // 新建用户模态框end

    // 设置模态框start
    const [settingData, setSettingData] = useState({});
    const [isSettingModalOpen, setIsSettingModalOpen] = useState(false);
    // 设置模态框end

    const fetchGroup = () => {

        fetch('/group-list', {
            method: 'GET',
        })
            .then(response => response.json())
            .then(d => {
                if (d.ok) {
                    let res = [];
                    for (let index = 0; index < d.data.length; index++) {
                        const element = d.data[index];
                        if (element.indexOf(keywords) !== -1) {
                            res.push(element);
                        }
                    }
                    setGroupList(res);
                } else {
                    messageApi.open({
                        type: 'error',
                        content: d.msg,
                    });
                }
            });
    };

    useEffect(fetchGroup, []);
    const newGroup = () => {
        fetch('/new-group', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(newGroupData)
        })
            .then(response => response.json())
            .then(d => {
                if (d.ok) {
                    setIsNewGroupModalOpen(false);
                    fetchGroup();
                } else {
                    messageApi.open({
                        type: 'error',
                        content: d.msg,
                    });
                }
            });
    };
    const newUser = () => {
        if (newUserData.ipassword !== newUserData.password) {
            messageApi.open({
                type: 'error',
                content: '密码不对应',
            });
            return;
        }
        fetch('/new-user', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(newUserData)
        })
            .then(response => response.json())
            .then(d => {
                if (d.ok) {
                    setIsNewUserModalOpen(false);
                    messageApi.open({
                        type: 'success',
                        content: '新增用户成功！',
                    });
                } else {
                    messageApi.open({
                        type: 'error',
                        content: d.msg,
                    });
                }
            });
    }
    const logout = () => {
        fetch('/logout', {
            method: 'GET'
        })
            .then(response => response.json())
            .then(d => {
                if (d.ok) {
                    Cookies.remove("session_id");
                    Cookies.remove("username");
                    navigate("/")
                } else {
                    messageApi.open({
                        type: 'error',
                        content: d.msg,
                    });
                }
            });
    };
    const setting = () => {
        if (settingData.inew !== settingData.new) {
            messageApi.open({
                type: 'error',
                content: '密码不对应',
            });
            return;
        }
        fetch('/user-password-update', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(settingData)
        })
            .then(response => response.json())
            .then(d => {
                if (d.ok) {
                    messageApi.open({
                        type: 'success',
                        content: '密码修改成功！',
                    });
                    logout();
                } else {
                    messageApi.open({
                        type: 'error',
                        content: d.msg,
                    });
                }
            });
    }
    return (
        <>
            {contextHolder}
            <Modal title="新建分组" open={isNewGroupModalOpen} onOk={newGroup} onCancel={() => {
                setIsNewGroupModalOpen(false);
            }}>
                <Input placeholder='分组名称' value={newGroupData.groupname} onChange={(e) => {
                    setNewGroupData({
                        groupname: e.target.value
                    })
                }}></Input>
            </Modal>
            <Modal title="新用户" open={isNewUserModalOpen} onOk={newUser} onCancel={() => {
                setIsNewUserModalOpen(false);
            }}>
                <Space direction="vertical" size="large">
                    <Input placeholder='用户名' value={newUserData.username} onChange={(e) => {
                        setNewUserData({
                            ...newUserData,
                            username: e.target.value
                        })
                    }}></Input>
                    <Input.Password placeholder='密码' value={newUserData.password} onChange={(e) => {
                        setNewUserData({
                            ...newUserData,
                            password: e.target.value
                        })
                    }} />
                    <Input.Password placeholder='确认密码' value={newUserData.ipassword} onChange={(e) => {
                        setNewUserData({
                            ...newUserData,
                            ipassword: e.target.value
                        })
                    }} />
                </Space>
            </Modal>
            <Modal title="设置" open={isSettingModalOpen} onOk={setting} onCancel={() => {
                setIsSettingModalOpen(false);
            }}>
                <Space direction="vertical" size="large">
                    <Input.Password placeholder='新密码' value={settingData.new} onChange={(e) => {
                        setSettingData({
                            ...settingData,
                            new: e.target.value
                        })
                    }} />
                    <Input.Password placeholder='确认密码' value={settingData.inew} onChange={(e) => {
                        setSettingData({
                            ...settingData,
                            inew: e.target.value
                        })
                    }} />
                    <Input.Password placeholder='原密码' value={settingData.old} onChange={(e) => {
                        setSettingData({
                            ...settingData,
                            old: e.target.value
                        })
                    }} />
                </Space>
            </Modal>
            <Layout className="layout">
                <Header className="header">
                    <Space>
                        用户主页
                        <Space.Compact style={{ width: '100%' }}>
                            <Input onKeyDown={(k) => {
                                if (k.code === "Enter") {
                                    fetchGroup();
                                }
                            }} value={keywords} placeholder="关键词" onChange={(e) => setKeywords(e.target.value)} />
                            <Button onClick={fetchGroup} icon={<SearchOutlined />}>搜索</Button>
                        </Space.Compact>
                        <Button icon={<AppstoreAddOutlined />} onClick={() => {
                            setIsNewGroupModalOpen(true);
                        }}>新建分组</Button>
                        <Button icon={<ExportOutlined />} onClick={() => {
                            window.open(`/export`);
                        }}>导出</Button>
                        <Button icon={<UserAddOutlined />} onClick={() => {
                            setIsNewUserModalOpen(true);
                        }}>新用户</Button>
                        <Button icon={<SettingOutlined />} onClick={() => {
                            setIsSettingModalOpen(true);
                        }}>设置</Button>
                        <Button icon={<LogoutOutlined />} onClick={logout}>登出</Button>
                    </Space>
                </Header>
                <Content className="content">
                    <Flex wrap gap="small">
                        {groupList.map(v => (
                            <Button key={v} onClick={() => {
                                navigate(`/group-main?groupname=${v}`);
                            }}>
                                {v}
                            </Button>
                        ))}
                    </Flex>
                </Content>
            </Layout>
        </>
    );
};


export default UserMain;
