<script setup>
    import { reactive } from 'vue'
    import { useRouter } from 'vue-router'

    const router = useRouter()

    
    const loginForm = reactive({
        username: '',
        password: ''
    });
    
    const handleLogin = () => {
        // 简单模拟登录验证
        if (loginForm.username && loginForm.password) {
            fetch(`/wmapi/login`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    username: loginForm.username,
                    password: loginForm.password
                })
            })
            .then(response => response.json())
            .then(d => {
                if (d.ok) {
                    localStorage.setItem('isLoggedIn', 'true');
                    router.push('/system');
                } else {
                    alert(d.msg);
                }
            });
        } else {
            alert('请输入用户名和密码');
        }
    };
</script>

<template>
<div class="login-container">
    <div class="login-form">
    <h2 class="login-title">Markdown 管理系统</h2>
    <form @submit.prevent="handleLogin">
        <div class="form-group">
        <label class="form-label">用户名</label>
        <input 
            type="text" 
            class="form-input" 
            v-model="loginForm.username" 
            placeholder="请输入用户名"
            required
        >
        </div>
        <div class="form-group">
        <label class="form-label">密码</label>
        <input 
            type="password" 
            class="form-input" 
            v-model="loginForm.password" 
            placeholder="请输入密码"
            required
        >
        </div>
        <button type="submit" class="login-btn">登录</button>
    </form>
    </div>
</div>
</template>

<style scoped>
    /* 登录页面样式 */
    .login-container {
      display: flex;
      flex-direction: column;
      justify-content: center;
      align-items: center;
      min-height: 100vh;
      padding: 20px;
      background: linear-gradient(135deg, #6a11cb 0%, #2575fc 100%);
    }
    
    .login-form {
      background: white;
      padding: 30px;
      border-radius: 12px;
      box-shadow: 0 10px 25px rgba(0, 0, 0, 0.2);
      width: 100%;
      max-width: 400px;
    }
    
    .login-title {
      text-align: center;
      margin-bottom: 30px;
      color: #333;
      font-size: 24px;
      font-weight: 600;
    }
        .form-group {
      margin-bottom: 20px;
    }
    
    .form-label {
      display: block;
      margin-bottom: 8px;
      font-weight: 500;
      color: #555;
    }
    
    .form-input {
      width: 100%;
      padding: 12px 15px;
      border: 1px solid #ddd;
      border-radius: 8px;
      font-size: 16px;
      transition: border-color 0.3s;
    }
    
    .form-input:focus {
      border-color: #6a11cb;
      outline: none;
    }
    
    .login-btn {
      width: 100%;
      padding: 12px;
      background: linear-gradient(135deg, #6a11cb 0%, #2575fc 100%);
      color: white;
      border: none;
      border-radius: 8px;
      font-size: 16px;
      font-weight: 600;
      cursor: pointer;
      transition: opacity 0.3s;
    }
    
    .login-btn:hover {
      opacity: 0.9;
    }
</style>
