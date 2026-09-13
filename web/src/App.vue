<template>
  <router-view />
</template>

<style>
:root {
  /* 参考小林简历配色：blue-600 主色 + slate 中性色 */
  --el-color-primary: #2563eb;
  --el-color-primary-light-3: #5b8def;
  --el-color-primary-light-5: #92b3f4;
  --el-color-primary-light-7: #c7d9f9;
  --el-color-primary-light-8: #dbe7fb;
  --el-color-primary-light-9: #ecf2fe;
  --el-color-primary-dark-2: #1d4fd7;
}
* { margin: 0; padding: 0; box-sizing: border-box; }
body {
  font-family: "IBM Plex Sans", "Noto Sans SC", "PingFang SC", "Microsoft YaHei",
    -apple-system, BlinkMacSystemFont, sans-serif;
  background: #f1f5f9; color: #0f172a;
}
</style>
