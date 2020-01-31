<!-- Implemented by Fengyu and Liangwei-->

<template>
  <div class="hello">
    <b-container class="bv-example-row">
  <b-row>
    <b-col>
      <label style="float: left;"> Elections </label>
      <div style="width: 50%">
          <select v-model="electionSelect" style="width:100%" multiple>
              <option v-for="(elec,index) in elections" :key="`elec-${index}`">
                {{ elec }}
              </option>
          </select>
        </div>
        <div style="float:left">
          <b-button class='float-center' style="margin-top: 5%" @click="updateInfo"> Update Election</b-button>
        </div>
        <div style='margin-top:60px'>
        <label> Public Key </label>
        <b-form>
          <p style='float:left'> Generator </p>
          <b-form-textarea v-model="publicKey['g']" debounce="500" rows="1" max-rows="1"></b-form-textarea>
          <p style='float:left'> Prime </p>
          <b-form-textarea v-model="publicKey['p']" debounce="500" rows="1" max-rows="1"></b-form-textarea>
          <p style='float:left'> ExponentPrime </p>
          <b-form-textarea v-model="publicKey['q']" debounce="500" rows="1" max-rows="1"></b-form-textarea>
          <p style='float:left'> PublicValue </p>
          <b-form-textarea v-model="publicKey['y']" debounce="500" rows="1" max-rows="1"></b-form-textarea>
        </b-form>
        </div>
    </b-col>
    <b-col>
    
    <label> Private Keys </label>
    <b-form v-for="(privatekey, index) in privateKeys" :key="index">
      <label> {{ privatekey['src'] }} </label>
      <b-form-textarea v-model="privatekey['privatekey']" debounce="500" rows="3" max-rows="5"></b-form-textarea>
    </b-form>
    
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


// Container IDs: ChatBox/NodeBox/PeerID
// Button IDS

export default {
  name: 'HelloWorld',
  props: {

  },
  data: function() {
    return {
      elections: [],
      electionSelect: [],
      publicKey: {},
      peersters: [],
      privateKeys: [],
      messages: [],
      keys: {},
    }
  },
  methods: {
    updateInfo: function() {
      var self = this;
      if (self.electionSelect.length == 0) {
        alert("Please select one election")
        return
      }

      var selectElection = self.electionSelect[0];

      self.publicKey = self.keys[selectElection]['publickey']

      self.privateKeys = self.keys[selectElection]['privatekeys']
    },
    pullMessage: async function() {
      var self = this
     var message = await fetch('/getElection', {method: 'GET', mode: 'cors'})
      .then(res => {
        if (res.ok) {
          var tmp = res.json();
          return tmp;
        }
      })
      .catch(err => {
        console.log(err)
      })

      if (message.messages.length == self.messages.length) {
        return
      }

      self.messages = message.messages

      self.elections = []

      self.messages.map(a => {
        self.elections.push(a['elec'])
      })

      self.keys = {}
      
      self.messages.map(a => {
        self.keys[a.elec] = a
      })
    },

    refresh: function() {
      this.pullMessage()
    }
  },
  mounted() {
    setInterval(this.refresh, 1000)
  }
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>

</style>


