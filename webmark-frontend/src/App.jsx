import { useState, lazy, Suspense } from 'react';
import { Layout, Input, Button, Typography } from 'antd';
import Cookies from 'js-cookie'; // 用于操作 cookie

import { useNavigate, BrowserRouter, Route, Routes } from 'react-router-dom';

const { Content } = Layout;
const { Title } = Typography;

const UserMain = lazy(() => import('./UserMain'));
const GroupMain = lazy(() => import('./GroupMain'));
const PublicHome = lazy(() => import('./PublicHome'));
const PublicDocViewer = lazy(() => import('./PublicDocViewer'));

// 登录页面组件
const LoginPage = () => {
    const navigate = useNavigate();
    const token = Cookies.get('session_id'); // 检查 cookie
    const [user, setUser] = useState({});

    // 登录逻辑
    const handleLogin = () => {
        // 这里可以添加你的登录逻辑，例如请求后端验证用户名和密码
        fetch('/wmapi/login', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(user)
        })
            .then(response => response.json())
            .then(d => {
                if (d.ok) {
                    navigate('/user-main');
                } else {
                    alert(d.msg);
                }
            });
    };

    return (
        <>
            {token ? <UserMain /> :
                <Layout>
                    <Content
                        style={{
                            display: 'flex',
                            justifyContent: 'center',
                            alignItems: 'center',
                            height: '100vh',
                        }}
                    >
                        <div
                            style={{
                                width: 300,
                                padding: '20px',
                                borderRadius: '8px',
                                boxShadow: '0 0 10px rgba(0, 0, 0, 0.1)',
                            }}
                        >
                            <Title level={3} style={{ textAlign: 'center' }}>
                                用户登录
                            </Title>
                            <Input
                                placeholder="请输入用户名"
                                value={user.username}
                                onChange={(e) => setUser({
                                    ...user,
                                    username: e.target.value
                                })}
                                style={{ marginBottom: '15px' }}
                            />
                            <Input.Password
                                placeholder="请输入密码"
                                value={user.password}
                                onChange={(e) => setUser({
                                    ...user,
                                    password: e.target.value
                                })}
                                style={{ marginBottom: '15px' }}
                            />
                            <Button type="primary" block onClick={handleLogin}>
                                登录
                            </Button>
                        </div>
                    </Content>
                </Layout>}
        </>
    );
};

// 公开文档查看器包装组件，用于从URL参数获取文档信息
const PublicDocViewerWrapper = ({ onBack }) => {
    const params = new URLSearchParams(window.location.search);
    const doc = {
        username: params.get('username'),
        groupname: params.get('groupname'),
        title: params.get('title')
    };

    if (!doc.username || !doc.groupname || !doc.title) {
        return (
            <Layout style={{ minHeight: '100vh' }}>
                <Content style={{ display: 'flex', justifyContent: 'center', alignItems: 'center' }}>
                    <div>文档不存在</div>
                </Content>
            </Layout>
        );
    }

    return <PublicDocViewer doc={doc} onBack={onBack} />;
};

// 路由入口组件，处理路由逻辑
const AppContent = () => {
    const navigate = useNavigate();
    const token = Cookies.get('session_id');

    const handleViewDoc = (doc) => {
        navigate(`/public-doc?username=${doc.username}&groupname=${doc.groupname}&title=${doc.title}`);
    };

    const handleBackToHome = () => {
        navigate('/');
    };

    const handleLogin = () => {
        navigate('/login');
    };

    return (
        <Routes>
            <Route path="/login" element={<LoginPage />} />
            <Route path="/" element={
                token ? <UserMain /> :
                    <Suspense fallback={<div>加载中...</div>}>
                        <PublicHome onViewDoc={handleViewDoc} onLogin={handleLogin} />
                    </Suspense>
            } />
            <Route path="/user-main" element={<Suspense> <UserMain /> </Suspense>} />
            <Route path='/group-main' element={<Suspense> <GroupMain /> </Suspense>} />
            <Route path="/public-doc" element={
                <Suspense fallback={<div>加载中...</div>}>
                    <PublicDocViewerWrapper onBack={handleBackToHome} />
                </Suspense>
            } />
        </Routes>
    );
};

// 主应用组件
const App = () => {
    return (<BrowserRouter>
        <AppContent />
    </BrowserRouter>)
};

export default App;