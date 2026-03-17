import { useState, useEffect } from 'react';
import { Layout, Input, Button, Card, Row, Col, Typography, Empty, Spin } from 'antd';
import { SearchOutlined, FileTextOutlined, UserOutlined } from '@ant-design/icons';

const { Header, Content } = Layout;
const { Title, Text } = Typography;

const PublicHome = ({ onViewDoc, onLogin }) => {
    const [keywords, setKeywords] = useState("");
    const [docList, setDocList] = useState([]);
    const [loading, setLoading] = useState(false);

    const fetchPublicDocs = (query = "") => {
        setLoading(true);
        const url = query
            ? '/wmapi/public-search'
            : '/wmapi/public-list';

        fetch(url, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ query })
        })
            .then(response => response.json())
            .then(d => {
                setLoading(false);
                if (d.ok) {
                    setDocList(d.data || []);
                } else {
                    setDocList([]);
                }
            })
            .catch(() => {
                setLoading(false);
                setDocList([]);
            });
    };

    useEffect(() => {
        fetchPublicDocs();
    }, []);

    const handleSearch = () => {
        fetchPublicDocs(keywords);
    };

    return (
        <Layout style={{ minHeight: '100vh' }}>
            <Header style={{
                background: '#fff',
                padding: '0 24px',
                borderBottom: '1px solid #f0f0f0',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'space-between'
            }}>
                <div style={{ display: 'flex', alignItems: 'center' }}>
                    <FileTextOutlined style={{ fontSize: 24, marginRight: 8, color: '#1677ff' }} />
                    <Title level={4} style={{ margin: 0 }}>Webmark</Title>
                </div>
                <Button type="primary" onClick={onLogin}>登录</Button>
            </Header>
            <Content style={{ padding: '24px 50px', background: '#f5f5f5' }}>
                <div style={{
                    maxWidth: 900,
                    margin: '0 auto'
                }}>
                    {/* 搜索区域 */}
                    <div style={{
                        background: '#fff',
                        padding: '24px',
                        borderRadius: 8,
                        marginBottom: 24,
                        boxShadow: '0 2px 8px rgba(0,0,0,0.06)'
                    }}>
                        <Input.Search
                            placeholder="搜索"
                            value={keywords}
                            onChange={(e) => setKeywords(e.target.value)}
                            onSearch={handleSearch}
                            enterButton={<><SearchOutlined />搜索</>}
                            size="large"
                            allowClear
                        />
                    </div>

                    {/* 文档列表 */}
                    <Spin spinning={loading}>
                        {docList.length > 0 ? (
                            <Row gutter={[16, 16]}>
                                {docList.map((doc, index) => (
                                    <Col xs={24} sm={12} lg={8} key={index}>
                                        <Card
                                            hoverable
                                            onClick={() => onViewDoc(doc)}
                                            style={{
                                                height: '100%',
                                            }}
                                            styles={{
                                                body: {
                                                    height: 100,
                                                    overflow: 'hidden'
                                                }
                                            }}
                                        >
                                            <Card.Meta
                                                avatar={<FileTextOutlined style={{ fontSize: 32, color: '#1677ff' }} />}
                                                title={
                                                    <Text strong ellipsis style={{ fontSize: 16 }}>
                                                        {doc.title}
                                                    </Text>
                                                }
                                                description={
                                                    <div>
                                                        <Text type="secondary" style={{ fontSize: 12 }}>
                                                            <UserOutlined /> {doc.username} · {doc.groupname}
                                                        </Text>
                                                    </div>
                                                }
                                            />
                                        </Card>
                                    </Col>
                                ))}
                            </Row>
                        ) : (
                            <Empty
                                description="暂无公开文档"
                                image={Empty.PRESENTED_IMAGE_SIMPLE}
                            />
                        )}
                    </Spin>
                </div>
            </Content>
        </Layout>
    );
};

export default PublicHome;