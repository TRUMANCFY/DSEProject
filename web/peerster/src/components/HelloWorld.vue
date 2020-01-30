  <template>
  <div class="hello">
    <b-container class="bv-example-row">
  <b-row>
    <b-col>
      <p style="text-align:left"> Election </p>
      <b-form-input v-model="election" placeholder="Enter election" style="width: 70%;"></b-form-input>
      <br>
      <p style="text-align:left"> Voter </p>
      <b-form-input v-model="voter" placeholder="Enter voter" style="width: 70%;"></b-form-input>
      <br>
      <b-button class='float-left' @click="getBlockChain" > Require Blockchain </b-button>
    </b-col>

    <b-col>
      <h4 style="text-align:left"> Blockchain </h4>
      <textarea class="form-control" id="ChatBox" rows="25" v-model="blockStr" readonly></textarea>
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
    }
  },
  methods: {
    getBlockChain: async function() {
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

      var fb = await fetch('/getblockchain', {method: "POST", body: JSON.stringify(payload), mode: 'cors'})
       .then(res => {
                if (res.ok) {
                    var tmp = res.json()
                    return tmp
                }
      })

      console.log(fb)
    },
  },
  mounted() {

  }
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>

</style>


