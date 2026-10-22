// Route paths and metadata recovered from the original frontend.
export const routes = [
  {
    path: "/install",
    component: () => import("@/views/install/Install.vue"),
    meta: {
      title: "系统安装",
    },
  },
  {
    path: "/home",
    component: () => import("@/views/home/Home.vue"),
    meta: {
      title: "首页",
      public: !0,
    },
  },
  {
    path: "/:entry/login",
    component: () => import("@/views/auth/Login.vue"),
    meta: {
      title: "登录",
    },
  },
  {
    path: "/:entry/app",
    component: () => import("@/layouts/OuterLayout.vue"),
    children: [
      {
        path: "",
        component: () => import("@/views/app/AppList.vue"),
        meta: {
          title: "应用列表",
        },
      },
    ],
  },
  {
    path: "/:entry/system",
    component: () => import("@/layouts/OuterLayout.vue"),
    children: [
      {
        path: "",
        component: () => import("@/views/system/SystemSettings.vue"),
        meta: {
          title: "系统设置",
          superOnly: true,
        },
      },
    ],
  },
  {
    path: "/:entry/docs",
    component: () => import("@/layouts/OuterLayout.vue"),
    children: [
      {
        path: "",
        component: () => import("@/views/docs/Documentation.vue"),
        meta: {
          title: "使用文档",
        },
      },
    ],
  },
  {
    path: "/:entry/admin",
    component: () => import("@/layouts/OuterLayout.vue"),
    children: [
      {
        path: "",
        component: () => import("@/views/admin/AdminList.vue"),
        meta: {
          title: "管理员",
          superOnly: true,
        },
      },
    ],
  },
  {
    path: "/:entry/profile",
    component: () => import("@/layouts/OuterLayout.vue"),
    children: [
      {
        path: "",
        component: () => import("@/views/profile/Profile.vue"),
        meta: {
          title: "个人中心",
        },
      },
    ],
  },
  {
    path: "/:entry",
    component: () => import("@/layouts/AdminLayout.vue"),
    children: [
      {
        path: "",
        redirect: (e) => `/${e.params.entry}/app`,
      },
      {
        path: "dashboard",
        component: () => import("@/views/dashboard/Dashboard.vue"),
        meta: {
          title: "总览",
          icon: "IconDashboard",
        },
      },
      {
        path: "user",
        component: () => import("@/views/user/UserList.vue"),
        meta: {
          title: "用户列表",
          group: "user",
          hideForTypes: ["kami"],
          submenu: {
            key: "user",
            title: "用户管理",
            icon: "IconUser",
          },
        },
      },
      {
        path: "user/group",
        component: () => import("@/views/user/UserGroup.vue"),
        meta: {
          title: "用户分组",
          group: "user",
          hideForTypes: ["kami"],
          submenu: {
            key: "user",
            title: "用户管理",
            icon: "IconUser",
          },
        },
      },
      {
        path: "user/edit/:id",
        component: () => import("@/views/user/UserEdit.vue"),
        meta: {
          title: "编辑用户",
          hidden: !0,
        },
      },
      {
        path: "app-info",
        component: () => import("@/views/app/AppInfo.vue"),
        meta: {
          title: "应用信息",
          superOnly: true,
          group: "app",
          icon: "IconApps",
        },
      },
      {
        path: "notice",
        component: () => import("@/views/app/Notice.vue"),
        meta: {
          title: "通知公告",
          group: "app",
          icon: "IconNotification",
        },
      },
      {
        path: "extend",
        component: () => import("@/views/app/Extend.vue"),
        meta: {
          title: "扩展配置",
          superOnly: true,
          group: "app",
          icon: "IconSettings",
        },
      },
      {
        path: "extend-file",
        component: () => import("@/views/app/ExtendFile.vue"),
        meta: {
          title: "扩展文件",
          superOnly: true,
          group: "app",
          icon: "IconFolder",
        },
      },
      {
        path: "version",
        component: () => import("@/views/app/Version.vue"),
        meta: {
          title: "版本控制",
          superOnly: true,
          group: "app",
          icon: "IconStorage",
        },
      },
      {
        path: "encrypt",
        component: () => import("@/views/app/Encrypt.vue"),
        meta: {
          title: "加密控制",
          superOnly: true,
          group: "app",
          icon: "IconSafe",
        },
      },
      {
        path: "blocklist",
        component: () => import("@/views/app/Blocklist.vue"),
        meta: {
          title: "黑名单",
          group: "app",
          icon: "IconStop",
        },
      },
      {
        path: "kami",
        component: () => import("@/views/kami/KamiList.vue"),
        meta: {
          title: "卡密列表",
          group: "user",
          submenu: {
            key: "kami",
            title: "卡密管理",
            icon: "IconBook",
          },
        },
      },
      {
        path: "kami/group",
        component: () => import("@/views/kami/KamiGroup.vue"),
        meta: {
          title: "卡密分组",
          group: "user",
          submenu: {
            key: "kami",
            title: "卡密管理",
            icon: "IconBook",
          },
        },
      },
      {
        path: "function",
        component: () => import("@/views/function/FunctionList.vue"),
        meta: {
          title: "云端函数",
          superOnly: true,
          group: "app",
          icon: "IconCode",
        },
      },
      {
        path: "function/edit/:id",
        component: () => import("@/views/function/FunctionEditor.vue"),
        meta: {
          title: "函数编辑",
          superOnly: true,
          hidden: !0,
        },
      },
      {
        path: "logs/user",
        component: () => import("@/views/logs/UserLogs.vue"),
        meta: {
          title: "用户日志",
          group: "app",
          hideForTypes: ["kami"],
          submenu: {
            key: "logs",
            title: "操作日志",
            icon: "IconFile",
          },
        },
      },
      {
        path: "logs/admin",
        component: () => import("@/views/logs/AdminLogs.vue"),
        meta: {
          title: "管理员日志",
          superOnly: true,
          group: "app",
          submenu: {
            key: "logs",
            title: "操作日志",
            icon: "IconFile",
          },
        },
      },
    ],
  },
  {
    path: "/",
    redirect: "/home",
  },
  {
    path: "/:pathMatch(.*)*",
    redirect: "/home",
  },
];
