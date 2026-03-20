import { useState, useEffect } from 'react';
import { Layout, Input, Button, Card, Row, Col, Typography, Empty, Spin, FloatButton } from 'antd';
import { SearchOutlined, FileTextOutlined, UserOutlined, EyeOutlined } from '@ant-design/icons';

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
        <Layout style={{ minHeight: '100vh', background: '#f5f5f5' }}>
            <Header style={{
                background: '#f5f5f5',
                padding: '0 24px',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'flex-end'
            }}>
                <Button
                    type="text"
                    icon={<UserOutlined />}
                    onClick={onLogin}
                    title="登录"
                />
            </Header>
            <Content style={{ padding: '24px', background: '#f5f5f5' }}>
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
                                                        <br />
                                                        <Text type="secondary" style={{ fontSize: 12 }}>
                                                            <EyeOutlined /> {doc.view_count || 0}
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
                <FloatButton.BackTop visibilityHeight={100} />
            </Content>
        </Layout>
    );
};

export default PublicHome;