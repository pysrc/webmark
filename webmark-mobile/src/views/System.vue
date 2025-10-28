<script setup>
    import { ref, onMounted } from 'vue'
    import { useRouter } from 'vue-router'
    import Cookies from 'js-cookie'
    const groups = ref([]);

    onMounted(async () => {
        fetch('/wmapi/group-list', {
            method: 'GET',
        })
        .then(response => response.json())
        .then(d => {
            if (d.ok) {
                for (let index = 0; index < d.data.length; index++) {
                    const element = d.data[index];
                    groups.value.push({
                        name: element,
                        icon: '📝',
                        count: 0
                    });
                }
            } else {
                alert(d.msg);
            }
        });
    });

    const router = useRouter();
    
    const viewGroup = (groupId) => {
        router.push(`/group/${groupId}`);
    };
    
    const handleLogout = () => {
        fetch('/wmapi/logout', {
            method: 'GET'
        })
        .then(response => response.json())
        .then(d => {
            if (d.ok) {
                Cookies.remove("session_id");
                Cookies.remove("username");
                localStorage.removeItem('isLoggedIn');
                router.push('/');
            } else {
                alert(d.msg);
            }
        });
    };
</script>

<template>
    <div class="system-container">
        <div class="header">
        <div class="header-title">Markdown 管理系统</div>
        <button class="logout-btn" @click="handleLogout">退出登录</button>
        </div>
        <div class="content">
        <h2 style="margin-bottom: 20px;">文档分组</h2>
        <div class="group-list">
            <div 
            v-for="group in groups" 
            :key="group.id" 
            class="group-item"
            @click="viewGroup(group.name)"
            >
            <div class="group-icon">{{ group.icon }}</div>
            <div class="group-name">{{ group.name }}</div>
            <div class="group-count">{{ group.count }} 个文档</div>
            </div>
        </div>
        </div>
    </div>
</template>

<style scoped>
    /* 系统页面样式 */
    .system-container {
      min-height: 100vh;
      display: flex;
      flex-direction: column;
    }
    
    .header {
      background: linear-gradient(135deg, #6a11cb 0%, #2575fc 100%);
      color: white;
      padding: 15px 20px;
      display: flex;
      justify-content: space-between;
      align-items: center;
      box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
    }
    
    .header-title {
      font-size: 20px;
      font-weight: 600;
    }
    .logout-btn {
      background: rgba(255, 255, 255, 0.2);
      color: white;
      border: none;
      padding: 8px 15px;
      border-radius: 20px;
      font-size: 14px;
      cursor: pointer;
    }
    
    .content {
      flex: 1;
      padding: 20px;
      overflow-y: auto;
    }
    /* 分组列表样式 */
    .group-list {
      display: grid;
      grid-template-columns: repeat(auto-fill, minmax(150px, 1fr));
      gap: 15px;
    }
    
    .group-item {
      background: white;
      border-radius: 12px;
      padding: 20px;
      box-shadow: 0 4px 8px rgba(0, 0, 0, 0.05);
      text-align: center;
      cursor: pointer;
      transition: transform 0.2s, box-shadow 0.2s;
    }
    
    .group-item:hover {
      transform: translateY(-5px);
      box-shadow: 0 8px 15px rgba(0, 0, 0, 0.1);
    }
    
    .group-icon {
      font-size: 32px;
      margin-bottom: 10px;
      color: #6a11cb;
    }
    
    .group-name {
      font-weight: 600;
      margin-bottom: 5px;
    }
    
    .group-count {
      font-size: 14px;
      color: #777;
    }
</style>