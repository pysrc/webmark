import { useState, useEffect } from 'react';
import { Layout, Button, Space, message, FloatButton } from 'antd';
import { Viewer } from '@bytemd/react';
import 'bytemd/dist/index.css';
import zhHans from 'bytemd/locales/zh_Hans.json';
import gfm from '@bytemd/plugin-gfm';
import gfmLocale from '@bytemd/plugin-gfm/locales/zh_Hans.json';
import highlight from '@bytemd/plugin-highlight';
import 'highlight.js/styles/vs.css';
import math from '@bytemd/plugin-math';
import mathLocale from '@bytemd/plugin-math/locales/zh_Hans.json';
import 'katex/dist/katex.css';
import mermaid from '@bytemd/plugin-mermaid';
import mermaidLocale from '@bytemd/plugin-mermaid/locales/zh_Hans.json';
import { visit } from 'unist-util-visit';

import {
    ArrowLeftOutlined,
    FileTextOutlined
} from '@ant-design/icons';

const { Header, Content } = Layout;

// 处理公开文档中的图片路径
const publicImagePrefix = (groupname, username) => {
    return {
        remark: (processor) =>
            processor.use(() => (tree) => {
                visit(tree, ['image', 'link'], (node) => {
                    if (typeof node.url === 'string' && !node.url.startsWith('http')) {
                        node.url = `/wmapi/public-markdown/${groupname}/${node.url.replace(/^\/+/, '')}`;
                    }
                });
            }),
    };
};

const PublicDocViewer = ({ doc, onBack }) => {
    const [content, setContent] = useState('');
    const [loading, setLoading] = useState(true);
    const [messageApi, contextHolder] = message.useMessage();

    // 构建plugins，使用doc的groupname和username处理图片路径
    const plugins = [
        gfm({ locale: gfmLocale }),
        highlight(),
        math({ locale: mathLocale, katexOptions: { output: 'mathml' } }),
        mermaid({ locale: mermaidLocale }),
        publicImagePrefix(doc?.groupname, doc?.username),
    ];

    useEffect(() => {
        if (doc && doc.username && doc.groupname && doc.title) {
            setLoading(true);
            fetch(`/wmapi/public-markdown/${doc.groupname}/${doc.title}.md`)
                .then(res => {
                    if (!res.ok) {
                        throw new Error('文档加载失败');
                    }
                    return res.text();
                })
                .then(text => {
                    setContent(text);
                    setLoading(false);
                })
                .catch(err => {
                    messageApi.open({
                        type: 'error',
                        content: '文档加载失败: ' + err.message,
                    });
                    setLoading(false);
                });
        }
    }, [doc]);

    if (!doc) {
        return null;
    }

    return (
        <>
            {contextHolder}
            <Layout style={{ minHeight: '100vh' }}>
                <Header style={{
                    position: 'fixed',
                    top: 0,
                    left: 0,
                    right: 0,
                    zIndex: 100,
                    background: '#fff',
                    padding: '0 24px',
                    borderBottom: '1px solid #f0f0f0',
                    display: 'flex',
                    alignItems: 'center'
                }}>
                    <Space>
                        <Button icon={<ArrowLeftOutlined />} onClick={onBack}>返回</Button>
                        <FileTextOutlined style={{ fontSize: 20, color: '#1677ff', marginLeft: 8 }} />
                        <span style={{ fontSize: 18, fontWeight: 500 }}>
                            {doc.title}
                        </span>
                        <span style={{ color: '#999', fontSize: 14 }}>
                            - {doc.username} · {doc.groupname}
                        </span>
                    </Space>
                </Header>
                <Content style={{ padding: '80px 50px 24px', background: '#fff' }}>
                    <div style={{
                        maxWidth: 900,
                        margin: '0 auto',
                        background: '#fff',
                        padding: '24px',
                        borderRadius: 8,
                        minHeight: 'calc(100vh - 150px)'
                    }}>
                        {loading ? (
                            <div style={{ textAlign: 'center', padding: '50px' }}>
                                加载中...
                            </div>
                        ) : (
                            <div className="markdown-body">
                                <Viewer
                                    value={content}
                                    plugins={plugins}
                                    locale={zhHans}
                                />
                            </div>
                        )}
                    </div>
                </Content>
                <FloatButton.BackTop visibilityHeight={100} />
            </Layout>
        </>
    );
};

export default PublicDocViewer;