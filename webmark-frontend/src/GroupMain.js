import { h } from 'hastscript';
import { useState, useEffect, useRef } from 'react';
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
import { visit } from 'unist-util-visit'

import CryptoJS from 'crypto-js';
import './GroupMain.css';

const headerStyle = {
    height: 60,
    paddingInline: 30,
    lineHeight: '30px',
    backgroundColor: '#fff',
};

const { Header, Content, Sider } = Layout;


const imagePrefix = (groupname) => {
    return {
        remark: (processor) =>
            processor.use(() => (tree) => {
                visit(tree, ['image', 'link'], (node) => {
                    if (typeof node.url === 'string' && !node.url.startsWith('http')) {
                        node.url = `/wmapi/markdown/${groupname}/${node.url.replace(/^\/+/, '')}`;
                    }
                });
            }),
    };
}

const uploadPlugin = ({ onUpload }) => {
    return {
        editorEffect: ({ editor }) => {
            const cm = editor
            const wrapper = cm.getWrapperElement()

            const prevent = (e) => {
                e.preventDefault()
                e.stopPropagation()
            }

            // 拖拽上传
            const handleDrop = async (e) => {
                prevent(e)
                const files = Array.from(e.dataTransfer.files)
                for (const file of files) {
                    const url = await onUpload(file)
                    let markdown =
                        file.type.startsWith('image/')
                            ? `![](${url})`
                            : `[${file.name}](${url})`
                    cm.replaceSelection(markdown + '\n')
                }
            }

            // 粘贴上传
            const handlePaste = async (e) => {
                const items = e.clipboardData.items
                if (!items) return

                for (const item of items) {
                    if (item.kind === 'file') {
                        const file = item.getAsFile()
                        if (!file) continue

                        const url = await onUpload(file)
                        let markdown =
                            file.type.startsWith('image/')
                                ? `![](${url})`
                                : `[${file.name}](${url})`
                        cm.replaceSelection(markdown + '\n')
                        e.preventDefault()
                    }
                }
            }

            wrapper.addEventListener('dragover', prevent)
            wrapper.addEventListener('drop', handleDrop)
            wrapper.addEventListener('paste', handlePaste)

            // 清理函数
            return () => {
                wrapper.removeEventListener('dragover', prevent)
                wrapper.removeEventListener('drop', handleDrop)
                wrapper.removeEventListener('paste', handlePaste)
            }
        },
    };
}

const codeCopyPlugin = () => {
    return {
        rehype: (processor) => {
            return processor.use(() => (tree) => {
                visit(tree, 'element', (node, index, parent) => {
                    // 情况1：代码块 <pre><code>
                    if (node.tagName === 'pre' && node.children?.[0]?.tagName === 'code') {
                        const codeNode = node.children[0];
                        const rawCode = codeNode.children.map((child) => child.value || '').join('');

                        node.children.push(
                            h(
                                'button',
                                {
                                    type: 'button',
                                    class: 'copy-btn',
                                    'data-code': rawCode,
                                    title: '复制代码',
                                },
                                '📋'
                            )
                        );
                        node.properties.className = (node.properties.className || []).concat('with-copy');
                    }

                    // 情况2：行内代码 <code>
                    if (node.tagName === 'code' && parent?.tagName !== 'pre') {
                        const rawCode = node.children.map((child) => child.value || '').join('');
                        // 包一层 span，让复制按钮能绝对定位
                        parent.children[index] = h(
                            'span',
                            { class: 'inline-code-wrapper' },
                            [
                                node, // 原始 <code>
                                h(
                                    'button',
                                    {
                                        type: 'button',
                                        class: 'inline-copy-btn',
                                        'data-code': rawCode,
                                        title: '复制代码',
                                    },
                                    '📋'
                                ),
                            ]
                        );
                    }
                });
            });
        },
    };
}


const GroupMain = () => {
    const [searchParams] = useSearchParams();
    const groupname = searchParams.get('groupname');
    const plugins = [
        gfm({
            locale: gfmLocale
        }),
        codeCopyPlugin(), // 代码复制插件
        highlight(),
        math({
            locale: mathLocale,
            katexOptions: { output: 'mathml' },
        }),
        mermaid({
            locale: mermaidLocale
        }),
        imagePrefix(groupname),
        uploadPlugin({
            onUpload: (file) => new Promise((resolve, _) => {
                if (!file) {
                    return;
                }
                // 创建FormData对象，用于将文件上传到服务器
                var formData = new FormData();
                // 将拖拽的文件添加到FormData对象中
                var name = `${Date.now()}_${file.name}`;
                formData.append('file', file, name);
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
                    resolve(`${mdname}/${name}`);
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
                xhr.open('POST', `/wmapi/upload/${groupname}/${mdname}`);
                xhr.send(formData);
            }),
        }),
    ]
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

    const textRef = useRef(mdvalue);
    const nameRef = useRef(mdname);
    const [isModified, setIsModified] = useState(false); // 标记是否有未保存的更改

    useEffect(() => {
        if(textRef.current !== mdvalue) {
            setIsModified(true);
        }
        textRef.current = mdvalue; // 每次渲染时更新 ref
    }, [mdvalue]);
    useEffect(() => {
        nameRef.current = mdname; // 每次渲染时更新 ref
    }, [mdname]);
    useEffect(() => {
        if (groupname && mdname) {
            document.title = `${groupname} - ${mdname}`;
        } else {
            document.title = 'WebMark'; // 默认标题
        }
    }, [groupname, mdname]);

    useEffect(() => {
        const handler = (e) => {
            const btn = e.target.closest('.copy-btn, .inline-copy-btn');
            if (btn) {
                const code = btn.getAttribute('data-code');
                
                // 使用现代 Clipboard API（优先）
                if (navigator.clipboard && window.isSecureContext) {
                    navigator.clipboard.writeText(code).then(() => {
                        // 切换成 ✅ 图标
                        btn.textContent = '✅';
                        setTimeout(() => {
                            btn.textContent = '📋';
                        }, 1500);
                    }).catch(err => {
                        console.error('Failed to copy text: ', err);
                    });
                } else {
                    // 降级处理：使用旧版 execCommand 方法
                    const textarea = document.createElement('textarea');
                    textarea.value = code;
                    document.body.appendChild(textarea);
                    textarea.select();
                    
                    try {
                        const successful = document.execCommand('copy');
                        if (successful) {
                            btn.textContent = '✅';
                            setTimeout(() => {
                                btn.textContent = '📋';
                            }, 1500);
                        } else {
                            console.error('Failed to copy text using execCommand');
                        }
                    } catch (err) {
                        console.error('Fallback copy failed: ', err);
                    }
                    
                    document.body.removeChild(textarea);
                }
            }
        };

        document.addEventListener('click', handler);
        return () => document.removeEventListener('click', handler);
    }, [mdvalue]);

    const fetchMarkdowns = () => {

        fetch('/wmapi/search-detail', {
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

    const saveMarkdown = () => {
        fetch(`/wmapi/update-markdown/${groupname}/${nameRef.current}`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/text'
            },
            body: textRef.current
        })
            .then(response => response.json())
            .then(d => {
                if (d.ok) {
                    // 通知保存成功
                    messageApi.open({
                        type: 'success',
                        content: '保存成功',
                    });
                    setIsModified(false);
                } else {
                    messageApi.open({
                        type: 'error',
                        content: d.msg,
                    });
                }
            });
    };
    const newMarkdown = () => {
        fetch(`/wmapi/new-markdown/${groupname}/${newMarkdownData.mdname}`, {
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
        if (!mdname) {
            return;
        }
        setMdName(mdname);
        fetch(`/wmapi/markdown/${groupname}/${mdname}.md?_t=${Date.now()}`, {
            method: 'GET',
            headers: {
                'Cache-Control': 'no-cache'
            }
        })
            .then(response => response.text())
            .then(d => {
                setIsModified(false);
                textRef.current = d;
                setMdValue(d);
                setShowEditor(true);
            });
    };
    const deleteConfirm = (e) => {
        setMdName("");
        fetch(`/wmapi/del-markdown/${groupname}/${mdname}`, {
            method: 'DELETE'
        })
            .then(response => response.text())
            .then(d => {
                setMdValue("");
                fetchMarkdowns();
            });
    };
    const deleteGroupConfirm = (e) => {
        fetch(`/wmapi/del-group/${groupname}`, {
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
    };
    const decrypt = () => {
        var _demd = CryptoJS.AES.decrypt(mdvalue, cryptoPwd).toString(CryptoJS.enc.Utf8);
        setMdValue(_demd);
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
                            window.open(`/wmapi/export/${groupname}`);
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
                                style={{
                                    height: '750px'
                                }}
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
                                            <Content><h1>{isModified ? '* ' : ''}{mdname}</h1></Content>
                                            <Sider width="40%" style={{ backgroundColor: '#fff' }}>
                                                <Space>
                                                    <Button icon={<SaveOutlined />} type="primary" onClick={saveMarkdown}>保存</Button>
                                                    <Button icon={<FileZipOutlined />} onClick={() => {
                                                        setIsCryptoModalOpen(true);
                                                    }}>加解密</Button>
                                                    <Button icon={<ExportOutlined />} onClick={() => {
                                                        window.open(`/wmapi/export/${groupname}/${mdname}`);
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
