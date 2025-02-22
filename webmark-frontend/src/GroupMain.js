import React, { useState, useEffect } from 'react';
import { Layout, Input, Button, Space, Modal, Splitter, List, message, Popconfirm } from 'antd';
import {
    AppstoreAddOutlined,
    ExportOutlined,
    SearchOutlined,
    DeleteOutlined,
    SaveOutlined,
    HomeOutlined,
    FileZipOutlined
} from '@ant-design/icons';

import { useSearchParams } from 'react-router-dom';
import Cookies from 'js-cookie';

import math from '@bytemd/plugin-math';
import mathLocale from '@bytemd/plugin-math/locales/zh_Hans.json';
import 'katex/dist/katex.css';

import highlight from '@bytemd/plugin-highlight';
import 'github-markdown-css';
import 'highlight.js/styles/vs.css';

import mermaid from '@bytemd/plugin-mermaid';
import mermaidLocale from '@bytemd/plugin-mermaid/locales/zh_Hans.json';

import { Editor } from '@bytemd/react';
import 'bytemd/dist/index.css';
import zhHans from 'bytemd/locales/zh_Hans.json';
import gfm from '@bytemd/plugin-gfm';
import gfmLocale from '@bytemd/plugin-gfm/locales/zh_Hans.json';
import './ComMain.css';

import CryptoJS from 'crypto-js';

const headerStyle = {
    height: 60,
    paddingInline: 30,
    lineHeight: '30px',
    backgroundColor: '#fff',
};

const { Header, Content, Sider } = Layout;

const plugins = [
    gfm({
        locale: gfmLocale
    }),
    highlight(),
    math({
        locale: mathLocale,
        katexOptions: { output: 'mathml' },
    }),
    mermaid({
        locale: mermaidLocale
    }),
]

const GroupMain = () => {
    const [searchParams] = useSearchParams();
    const groupname = searchParams.get('groupname');
    const [keywords, setKeywords] = useState("");
    const [markdownList, setMarkdownList] = useState([]);
    // 新建分组模态框start
    const [newMarkdownData, setNewMarkdownData] = useState({});
    const [isNewMarkdownModalOpen, setIsNewMarkdownModalOpen] = useState(false);
    // 新建分组模态框end

    // 加解密模态框start
    const [cryptoPwd, setCryptoPwd] = useState("");
    const [isCryptoModalOpen, setIsCryptoModalOpen] = useState(false);
    // 加解密模态框end
    const [mdvalue, setMdValue] = useState('');
    const [mdname, setMdName] = useState('');
    const [showEditor, setShowEditor] = useState(false);
    const [messageApi, contextHolder] = message.useMessage();

    const fetchMarkdowns = () => {

        fetch('/search-detail', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                group: groupname,
                query: keywords
            })
        })
            .then(response => response.json())
            .then(d => {
                if (d.ok) {
                    setMarkdownList(d.data);
                    if (d.data) {
                        openMarkdown(d.data[0])();
                    }
                } else {
                    messageApi.open({
                        type: 'error',
                        content: d.msg,
                    });
                }
            });
    };

    useEffect(fetchMarkdowns, []);

    useEffect(() => {
        Cookies.set("groupname", groupname);
    }, []);

    const saveMarkdown = () => {
        const mdname = localStorage.getItem("mdname");
        const mdvalues = localStorage.getItem("mdvalue");
        fetch(`/update-markdown/${groupname}/${mdname}`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/text'
            },
            body: mdvalues
        })
            .then(response => response.json())
            .then(d => {
                if (d.ok) {
                    // 通知保存成功
                    messageApi.open({
                        type: 'success',
                        content: '保存成功',
                    });
                } else {
                    messageApi.open({
                        type: 'error',
                        content: d.msg,
                    });
                }
            });
    };
    const newMarkdown = () => {
        fetch(`/new-markdown/${groupname}/${newMarkdownData.mdname}`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/text'
            },
            body: `## 操作介绍
* Ctrl+S保存
* 图片支持复制粘贴、拖拽`
        })
            .then(response => response.json())
            .then(d => {
                if (d.ok) {
                    fetchMarkdowns();
                    setIsNewMarkdownModalOpen(false);
                } else {
                    messageApi.open({
                        type: 'error',
                        content: d.msg,
                    });
                }
            });
    }
    const openMarkdown = (mdname) => () => {
        if(!mdname) {
            return;
        }
        localStorage.setItem("mdname", mdname);
        setMdName(mdname);
        fetch(`/${mdname}.md`, {
            method: 'GET',
            headers: {
                'Cache-Control': 'no-cache'
            }
        })
            .then(response => response.text())
            .then(d => {
                setMdValue(d);
                setShowEditor(true);
            });
    };
    const deleteConfirm = (e) => {
        const mdname = localStorage.getItem("mdname");
        localStorage.removeItem("mdname");
        setMdName("");
        fetch(`/del-markdown/${groupname}/${mdname}`, {
            method: 'DELETE'
        })
            .then(response => response.text())
            .then(d => {
                setMdValue("");
                fetchMarkdowns();
            });
    };
    const deleteGroupConfirm = (e) => {
        fetch(`/del-group/${groupname}`, {
            method: 'DELETE'
        })
            .then(response => response.text())
            .then(d => {
                window.location.href = "/";
            });
    };
    const encrypt = () => {
        var _enmd = CryptoJS.AES.encrypt(mdvalue, cryptoPwd).toString();
        setMdValue(_enmd);
        localStorage.setItem("mdvalue", _enmd);
    };
    const decrypt = () => {
        var _demd = CryptoJS.AES.decrypt(mdvalue, cryptoPwd).toString(CryptoJS.enc.Utf8);
        setMdValue(_demd);
        localStorage.setItem("mdvalue", _demd);
    };
    return (
        <>
            {contextHolder}
            <Modal title="新建文档" open={isNewMarkdownModalOpen} onOk={newMarkdown} onCancel={() => {
                setIsNewMarkdownModalOpen(false);
            }}>
                <Input placeholder='文档名称' value={newMarkdownData.mdname} onChange={(e) => {
                    setNewMarkdownData({
                        mdname: e.target.value
                    })
                }}></Input>
            </Modal>
            <Modal title="加解密" open={isCryptoModalOpen} onOk={() => {
                setIsCryptoModalOpen(false);
            }} onCancel={() => {
                setIsCryptoModalOpen(false);
            }}>
                <Space direction="vertical">
                    <Input.Password placeholder='密码' value={cryptoPwd} onChange={(e) => {
                        setCryptoPwd(e.target.value);
                    }} />
                    <Space>
                        <Button type="primary" onClick={encrypt}>加密</Button>
                        <Button onClick={decrypt}>解密</Button>
                    </Space>
                </Space>
            </Modal>
            <Layout className="layout">
                <Header className="header">
                    <Space>
                        {groupname}
                        <Button icon={<HomeOutlined />} onClick={() => {
                            window.location.href = "/";
                        }}>首页</Button>
                        <Space.Compact style={{ width: '100%' }}>
                            <Input onKeyDown={(k) => {
                                if (k.code === "Enter") {
                                    fetchMarkdowns();
                                }
                            }} value={keywords} placeholder="关键词" onChange={(e) => setKeywords(e.target.value)} />
                            <Button onClick={fetchMarkdowns} icon={<SearchOutlined />}>搜索</Button>
                        </Space.Compact>
                        <Button icon={<AppstoreAddOutlined />} onClick={() => {
                            setIsNewMarkdownModalOpen(true);
                        }}>新建文档</Button>
                        <Button icon={<ExportOutlined />} onClick={() => {
                            window.open(`/export/${groupname}`);
                        }}>导出分组</Button>
                        <Popconfirm
                            title="删除分组"
                            description="确认删除分组？"
                            onConfirm={deleteGroupConfirm}
                            onCancel={(e) => { }}
                            okText="是"
                            cancelText="否"
                        >
                            <Button danger icon={<DeleteOutlined />}>删除分组</Button>
                        </Popconfirm>

                    </Space>
                </Header>
                <Content className="content">
                    <Splitter>
                        <Splitter.Panel defaultSize="15%" max="70%">
                            <List
                                dataSource={markdownList}
                                renderItem={(item) => (
                                    <List.Item>
                                        <Button type='link' onClick={openMarkdown(item)}>{item}</Button>
                                    </List.Item>
                                )}
                            />
                        </Splitter.Panel>
                        <Splitter.Panel>
                            {showEditor ?
                                <Layout>
                                    <Header style={headerStyle}>
                                        <Layout style={{ backgroundColor: '#fff' }}>
                                            <Content><h1>{mdname}</h1></Content>
                                            <Sider width="40%" style={{ backgroundColor: '#fff' }}>
                                                <Space>
                                                    <Button icon={<SaveOutlined />} type="primary" onClick={saveMarkdown}>保存</Button>
                                                    <Button icon={<FileZipOutlined />} onClick={() => {
                                                        setIsCryptoModalOpen(true);
                                                    }}>加解密</Button>
                                                    <Button icon={<ExportOutlined />} onClick={() => {
                                                        window.open(`/export/${groupname}/${mdname}`);
                                                    }}>导出</Button>

                                                    <Popconfirm
                                                        title="删除文章"
                                                        description="确认删除文档？"
                                                        onConfirm={deleteConfirm}
                                                        onCancel={(e) => { }}
                                                        okText="是"
                                                        cancelText="否"
                                                    >
                                                        <Button danger icon={<DeleteOutlined />}>删除</Button>
                                                    </Popconfirm>
                                                </Space>
                                            </Sider>
                                        </Layout>
                                    </Header>
                                    <Content>
                                        <Editor
                                            value={mdvalue}
                                            plugins={plugins}
                                            onChange={(v) => {
                                                setMdValue(v);
                                                localStorage.setItem("mdvalue", v);
                                            }}
                                            editorConfig={{
                                                lineNumbers: true,
                                                autofocus: false,
                                                extraKeys: {
                                                    "Ctrl-S": () => {
                                                        // 保存
                                                        saveMarkdown();
                                                        return true; // 返回 true 表示事件已被处理
                                                    }
                                                },
                                            }}
                                            locale={zhHans}
                                            uploadImages={(files) => new Promise((resolve, _) => {
                                                if (!files) {
                                                    return;
                                                }
                                                const mdname = localStorage.getItem("mdname");
                                                // 创建FormData对象，用于将文件上传到服务器
                                                var formData = new FormData();
                                                var image_names = [];
                                                // var file_names = [];
                                                // 将拖拽的文件添加到FormData对象中
                                                for (var i = 0; i < files.length; i++) {
                                                    var name = `${Date.now()}_${files[i].name}`;
                                                    formData.append('file', files[i], name);
                                                    if (files[i].type.indexOf('image') !== -1) {
                                                        image_names.push(name);
                                                    }
                                                    // else {
                                                    //     file_names.push(name);
                                                    // }
                                                }
                                                // 创建XMLHttpRequest对象，用于发送请求
                                                var xhr = new XMLHttpRequest();
                                                // 监听上传进度
                                                xhr.upload.addEventListener('progress', (e) => {
                                                    if (e.lengthComputable) {
                                                        var percent = (e.loaded / e.total) * 100;
                                                        percent = parseInt(percent);
                                                        console.log('上传进度：' + percent + '%');
                                                    }
                                                });
                                                // 监听上传完成事件
                                                xhr.addEventListener('load', (e) => {
                                                    console.log('上传完成');
                                                    let ri = [];
                                                    for (var i = 0; i < image_names.length; i++) {
                                                        ri.push({
                                                            url: `${mdname}/${image_names[i]}`
                                                        });
                                                    }

                                                    // for (var i = 0; i < file_names.length; i++) {
                                                    //     ri.push({
                                                    //         url: `${mdname}/${file_names[i]}`
                                                    //     });
                                                    // }
                                                    resolve(ri);
                                                });
                                                // 监听上传出错事件
                                                xhr.addEventListener('error', (e) => {
                                                    console.log('上传出错');
                                                });
                                                // 监听上传取消事件
                                                xhr.addEventListener('abort', (e) => {
                                                    console.log('上传取消');
                                                });
                                                // 发送请求
                                                xhr.open('POST', `/upload/${groupname}/${mdname}`);
                                                xhr.send(formData);
                                            })}
                                        />
                                    </Content>
                                </Layout> : <></>}
                        </Splitter.Panel>
                    </Splitter>
                </Content>
            </Layout>
        </>
    );
};


export default GroupMain;
