import { useState, lazy, Suspense } from 'react';
import { Layout, Input, Button, Typography } from 'antd';
import Cookies from 'js-cookie'; // 用于操作 cookie

import { useNavigate, BrowserRouter, Route, Routes } from 'react-router-dom';

const { Content } = Layout;
const { Title } = Typography;

const UserMain = lazy(() => import('./UserMain'));
const GroupMain = lazy(() => import('./GroupMain'));

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

// 主应用组件
const App = () => {
    return (<BrowserRouter>
        <Routes>
            <Route path="/" element={<LoginPage />} />
            <Route path="/user-main" element={<Suspense> <UserMain /> </Suspense>} />
            <Route path='/group-main' element={<Suspense> <GroupMain /> </Suspense>} />
        </Routes>
    </BrowserRouter>)
};

export default App;