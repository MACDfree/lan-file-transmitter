<template>
<el-container>
  <el-header>
    <el-input v-model="path" placeholder="请输入本地文件路径">
      <el-button slot="append" icon="el-icon-plus" @click="addPath"></el-button>
    </el-input>
  </el-header>
  <el-main>
    <el-table
      :data="urls"
      style="width: 100%">
      <el-table-column
        prop="path"
        label="本地路径"
        min-width="400">
      </el-table-column>
      <el-table-column
        label="下载地址"
        width="100"
        align="center">
        <template slot-scope="scope">
          <a :href="scope.row.url">下载</a>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="60">
        <template slot-scope="scope">
          <el-button
            type="text"
            @click="handleDelete(scope.row.url)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>
    </el-main>
</el-container>
</template>

<script>
import axios from "axios"

export default {
  name: "Uploader",
  props: {
    msg: String
  },
  data() {
    return {
      path: "",
      urls: []
    };
  },
  mounted() {
    axios.get("/rest/list").then(res=>{
      this.urls = res.data.downloadUrls
    })
  },
  methods: {
  addPath() {
      axios.post("/rest/upload", {path: this.path}).then(res=>{
        this.urls = res.data.downloadUrls
        this.path = ""
      }).catch(err=>{
        this.$message.error('错误信息：'+err.response.data.err);
      });
    },
    handleDelete(url) {
      let uuid = url.substring(url.lastIndexOf("/")+1)
      axios.post("/rest/delete", {id: uuid}).then(res=>{
        this.$message({
          message: '删除成功',
          type: 'success'
        })
        this.urls = res.data.downloadUrls
      }).catch(err=>{
          this.$message.error('错误信息：'+err.response.data.err);
        })
    }
  }
};
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>
</style>
