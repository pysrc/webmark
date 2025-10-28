<script setup>
    import { ref, onMounted } from 'vue'
    import { useRouter, useRoute } from 'vue-router'
    import Cookies from 'js-cookie'

    const route = useRoute();
    const router = useRouter();
    
    const groupname = route.params.groupname;
    const searchKeyword = ref('');
    const documents = ref([]);

    const fetchList = () => {
        fetch('/wmapi/search-detail', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                group: route.params.groupname,
                query: searchKeyword.value
            })
        })
            .then(response => response.json())
            .then(d => {
                if (d.ok) {
                    if (d.data) {
                        documents.value = d.data;
                    }
                } else {
                    messageApi.open({
                        type: 'error',
                        content: d.msg,
                    });
                }
            });
    }

    onMounted(async () => {
        fetchList();
    });
    
    const goBack = () => {
        router.push('/system');
    };
    
    const viewDocument = (doc) => {
        let n = `/document/${route.params.groupname}/${doc}`;
        router.push(n);
    };
    
    // 新建文档
    const createNewDocument = () => {
        router.push(`/newdocument/${route.params.groupname}/title`);
    };
    
    const clearSearch = () => {
        searchKeyword.value = '';
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
        <div class="header-title">文档分组</div>
        <button class="logout-btn" @click="handleLogout">退出登录</button>
        </div>
        <div class="content">
        <button class="back-btn" @click="goBack">
            <span>←</span> 返回分组
        </button>
        <div class="header-actions">
            <h2 style="margin-bottom: 20px;">{{ groupname }}</h2>
            <button class="new-doc-btn" @click="createNewDocument">
                + 新建文档
            </button>
        </div>
        
        <!-- 搜索框 -->
        <div class="search-container">
            <input 
            type="text" 
            class="search-input" 
            v-model="searchKeyword"
            placeholder="搜索文档标题..."
            >
            <button v-if="searchKeyword" class="search-clear" @click="clearSearch">×</button>
            <div class="search-icon" @click="fetchList">🔍</div>
        </div>
        
        <!-- 文档列表 -->
        <div v-if="documents.length > 0" class="doc-list">
            <div 
            v-for="doc in documents" 
            :key="doc" 
            class="doc-item"
            @click="viewDocument(doc)"
            >
            <div class="doc-info">
                <div class="doc-title">{{ doc }}</div>
            </div>
            <div class="doc-arrow">→</div>
            </div>
        </div>
        
        <!-- 空状态 -->
        <div v-else class="empty-state">
            <div class="empty-icon">📄</div>
            <div class="empty-text">
            {{ searchKeyword ? '没有找到相关文档' : '该分组暂无文档' }}
            </div>
            <button class="new-doc-btn empty-btn" @click="createNewDocument">
                + 新建文档
            </button>
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
    .back-btn {
      background: #f0f0f0;
      border: none;
      padding: 10px 15px;
      border-radius: 8px;
      margin-bottom: 20px;
      cursor: pointer;
      display: flex;
      align-items: center;
      gap: 5px;
    }
    
    /* 头部操作区域 */
    .header-actions {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 20px;
    }
    
    .new-doc-btn {
      background: linear-gradient(135deg, #6a11cb 0%, #2575fc 100%);
      color: white;
      border: none;
      padding: 10px 20px;
      border-radius: 8px;
      font-size: 14px;
      cursor: pointer;
      transition: all 0.3s;
      box-shadow: 0 2px 8px rgba(106, 17, 203, 0.3);
    }
    
    .new-doc-btn:hover {
      transform: translateY(-2px);
      box-shadow: 0 4px 12px rgba(106, 17, 203, 0.4);
    }
    
    .empty-btn {
      margin-top: 20px;
    }
    
    /* 搜索框样式 */
    .search-container {
      margin-bottom: 20px;
      position: relative;
    }
    
    .search-input {
      width: 85%;
      padding: 12px 45px 12px 15px;
      border: 1px solid #ddd;
      border-radius: 25px;
      font-size: 16px;
      background-color: #f9f9f9;
      transition: all 0.3s;
    }
    
    .search-input:focus {
      border-color: #6a11cb;
      background-color: white;
      outline: none;
      box-shadow: 0 0 0 3px rgba(106, 17, 203, 0.1);
    }
    
    .search-icon {
      position: absolute;
      right: 15px;
      top: 50%;
      transform: translateY(-50%);
      color: #777;
    }
    
    .search-clear {
      position: absolute;
      right: 40px;
      top: 50%;
      transform: translateY(-50%);
      background: #ddd;
      border: none;
      width: 20px;
      height: 20px;
      border-radius: 50%;
      display: flex;
      align-items: center;
      justify-content: center;
      font-size: 12px;
      cursor: pointer;
      color: #555;
    }
    .doc-list {
      background: white;
      border-radius: 12px;
      overflow: hidden;
      box-shadow: 0 4px 8px rgba(0, 0, 0, 0.05);
    }
    
    .doc-item {
      padding: 15px 20px;
      border-bottom: 1px solid #f0f0f0;
      cursor: pointer;
      display: flex;
      justify-content: space-between;
      align-items: center;
      transition: background-color 0.2s;
    }
    
    .doc-item:last-child {
      border-bottom: none;
    }
    
    .doc-item:hover {
      background-color: #f9f9f9;
    }
    
    .doc-info {
      flex: 1;
    }
    
    .doc-title {
      font-weight: 600;
      margin-bottom: 5px;
    }
    
    .doc-date {
      font-size: 14px;
      color: #777;
    }
    
    .doc-arrow {
      color: #999;
    }
    
    /* 文档详情样式 */
    .doc-detail {
      background: white;
      border-radius: 12px;
      padding: 20px;
      box-shadow: 0 4px 8px rgba(0, 0, 0, 0.05);
    }
    
    .doc-header {
      margin-bottom: 20px;
      padding-bottom: 15px;
      border-bottom: 1px solid #f0f0f0;
      display: flex;
      justify-content: space-between;
      align-items: flex-start;
    }
    
    .doc-title-section {
      flex: 1;
    }
    
    .doc-detail-title {
      font-size: 22px;
      font-weight: 600;
      margin-bottom: 10px;
    }
    
    .doc-meta {
      display: flex;
      gap: 15px;
      font-size: 14px;
      color: #777;
    }
    
    .doc-actions {
      display: flex;
      gap: 10px;
    }
        /* 空状态样式 */
    .empty-state {
      text-align: center;
      padding: 40px 20px;
      color: #777;
    }
    
    .empty-icon {
      font-size: 48px;
      margin-bottom: 15px;
      opacity: 0.5;
    }
    
    .empty-text {
      font-size: 16px;
    }
</style>