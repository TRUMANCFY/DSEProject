<!-- Implemented by Fengyu and Liangwei-->

<template>
  <div class="hello">
    <b-container class="bv-example-row">
  <b-row>
    <b-col>
      <label> Peerster: {{peerster}} </label>
      <p style="text-align:left"> Election </p>
      <b-form-input v-model="election" placeholder="Enter election" style="width: 70%;"></b-form-input>
      <br>
      <p style="text-align:left"> Voter </p>
      <b-form-input v-model="voter" placeholder="Enter voter" style="width: 70%;"></b-form-input>
      <br>
      <b-button class='float-left' @click="postBlockChain" > Require Blockchain </b-button>
    </b-col>

    <b-col>
      <h4 style="text-align:left"> Blockchain </h4>
      <textarea class="form-control" id="ChatBox" rows="13" v-model="blockStr" readonly></textarea>
      <h4 style="text-align:left"> Error </h4>
      <textarea class="form-control" id="ChatBox" rows="13" v-model="errStr" readonly></textarea>
    </b-col>
  </b-row>
</b-container>
  </div>
</template>

<script>
import Vue from 'vue'
import BootstrapVue from 'bootstrap-vue'

import 'bootstrap/dist/css/bootstrap.css'
import 'bootstrap-vue/dist/bootstrap-vue.css'

Vue.use(BootstrapVue)

export default {
  name: 'HelloWorld',
  props: {

  },
  data: function() {
    return {
      election: "",
      voter: "",
      blockStr: "",
      errStr: "",
      blockMsg: [],
      errMsg: [],
      peerster: "",
    }
  },
  methods: {
    postBlockChain: async function() {
      var self = this;

      if (self.election == "") {
        alert("Please provide the election")
        return
      }

      if (self.voter == "") {
        alert("Please provide the voter")
        return
      }

      var payload = {
          'election': self.election,
          'voter': self.voter,
      }

      var fb = await fetch('/postblockchain', {method: "POST", body: JSON.stringify(payload), mode: 'cors'})
       .then(res => {
                if (res.ok) {
                    var tmp = res.json()
                    return tmp
                }
      })

      console.log(fb)
    },

    getBlockchain: async function() {
      var self = this;

      var a = await fetch('/getblockchain', {method: "GET", mode: 'cors'})
      .then(res => {
        if (res.ok) {
          var tmp = res.json()
          return tmp
        }
      })

      var messageFilter = []
      var errorFilter = []

      a['blocks'].map(d => {
        if (d.startsWith("ERROR")) {
          errorFilter.push(d)
        } else {
          messageFilter.push(d)
        }
      })
      
      if (self.blockMsg.length != messageFilter.length) {
        self.blockMsg = messageFilter;
        self.blockStr = self.blockMsg.join('\n')
      }
      
      if (self.errMsg.length != errorFilter.length) {
        self.errMsg = errorFilter;
        self.errStr = self.errMsg.join('\n')
        //self.errStr = self.errMsg[0];
      }

    }
  },
  async mounted() {
    var self = this;

    var peersterJson = await fetch('/id', {method: "GET", mode: "cors"})
    .then(res => {
      if (res.ok) {
        var tmp = res.json()
        return tmp;
      }
    })

    self.peerster = peersterJson.id

    setInterval(this.getBlockchain, 100)
  }
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>

</style>


