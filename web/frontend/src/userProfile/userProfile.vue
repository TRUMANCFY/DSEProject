<!-- Implemented by Fengyu -->

<template>
    <div>
        <h2>User Profile</h2>
        <div id='profile' style='margin-top:20px;'>
            <div id='leftProfile'>
                <img v-bind:src="registeredImage" alt="Avatar" class="avatar">
            </div>

            <div id='rightProfile'>
                <table style='float:left'>
                    <tr>
                    <th> First Name: </th>
                    <td>{{ user.firstName }}</td>
                    </tr>
                    <tr>
                        <th> Last Name: </th>
                        <td>{{ user.lastName }}</td>
                    </tr>
                    <tr>
                        <th> Email Address: </th>
                        <td>{{ user.email }}</td>
                    </tr>
                </table>
            </div>
        </div>


        <button class="btn btn-primary" @click='createElection' style='margin-top:10px;'> Create an election </button>

        <div class='form-group' style='margin-top: 10px;'>
            <h5> Participate an election </h5>
            <b-form-select v-model="selected" :options="elections" :select-size="6" style='width: 40%;'></b-form-select>
            <button class="btn btn-primary" @click="goVote" style='margin-left:10px'>Go to vote</button>
        </div>

        <div class='form-group' style='margin-top: 10px;'>
            <h5> End an election </h5>
            <b-form-select v-model="selectedCreate" :options="electionCreate" :select-size="2" style='width: 40%;'></b-form-select>
            <button class="btn btn-primary" @click="endElection" style='margin-left:10px'> End </button>
        </div>

         <div class='form-group' style='margin-top: 10px;'>
            <h5> View election result </h5>
            <b-form-select v-model="selectedResult" :options="electionsFinish" :select-size="3" style='width:40%;'></b-form-select>
            <button class="btn btn-primary" @click="viewResult" style='margin-left:10px'> View </button>
        </div>
    </div>
</template>

<script>
import { mapState, mapActions } from 'vuex'
import config from 'config'

// import image
import registeredImage from '../assets/registeredUser.png'

export default {
    data () {
        return {
            registeredImage: registeredImage,
            elections: [],
            selected: null,
            selectedResult: null,
            selectedCreate: null,
            electionsFinish: [],
            electionCreate: [],
        }
    },
    computed: {
        ...mapState('account', ['status', 'user'])
    },
    methods: {
        ...mapActions('account', ['register']),
        getImage: function() {
            return this.registeredImage
        },
        goVote: function() {
            var self = this;

            // check the box
            if (self.selected == null) {
                alert('Please select one election')
                return
            }

            var election = {}

            self.elections.map(e => {
                if (e['name'] == self.selected) {
                    election = e;
                }
            })

            this.$router.push({ name: 'vote', path: '/vote', params: { questions: election }})
        },
        getElection: async function() {
            var self = this;

            var newElections = await fetch(`${config.apiUrl}/getElection`, {method: 'GET', mode: 'cors'})
            .then(res => {
                if (res.ok) {
                    var tmp = res.json();
                    return tmp;
                }
            });

            var electionCreate = [];
            
            newElections.map(e => {
                if (e.creator == self.user.id) {
                    electionCreate.push(e.name)
                }
            })

            if (electionCreate.length != self.electionCreate.length) {
                self.electionCreate = electionCreate;
            }
            

            if (newElections.length != self.elections.length) {
                self.elections = newElections;
                
                self.elections.forEach(element => {
                    element['text'] = element['name'];
                    element['value'] = element['name'];
                });
            }

        },
        createElection: function() {
            this.$router.push('/create')
        },
        getElectionFinish: async function() {
            var self = this;
            
            var payload = {
                'voter': self.user.id,
            }
            var electionDone = await fetch(`${config.apiUrl}/getVoted`, {method: 'POST', body: JSON.stringify(payload), mode: 'cors'})
            .then(res => {
                if (res.ok) {
                    var tmp = res.json();
                    return tmp;
                }
            });

            var electionsF = []

            electionDone.map(a => {
                electionsF.push(a.election)
            })

            if (electionsF.length != self.electionsFinish.length) {
                self.electionsFinish = electionsF;
            }
        },
        endElection: async function() {
            console.log('end')
            var self = this;

            if (self.selectedCreate == null) {
                alert("Please select vote to end")
                return
            }

            console.log(self.selectedCreate)

            var payload = {
                'electionend': self.selectedCreate,
            }

            var electionEnd = await fetch('/endvote', {method: 'POST', body: JSON.stringify(payload), mode: 'cors'})
            .then(res => {
                if (res.ok) {
                    return '';
                }
            });
        },
        viewResult: async function() {

            var self = this;

            if (self.selectedResult == null) {
                alert("Please select vote to view")
                return
            }

            // try to get the voting result first
            var payload = {
                'elec': self.selectedResult,
            }

            var result = await fetch('http://127.0.0.1:8082/getresult', {method: 'POST', body: JSON.stringify(payload), mode: 'cors'})
            .then(res => {
                if (res.ok) {
                    return res.json();
                }
            });

            console.log(result);

            if (!result['exist']) {
                alert("The result is still collecting!")
                return
            }

            // push to router
            // this.$router.push()
            var election = {}

            self.elections.map(e => {
                if (e['name'] == self.selectedResult) {
                    election = e;
                }
            })

            console.log(election)

            this.$router.push({ name: 'result', path: '/result', params: { questionsIn: election, result: result }})
        }
    },
    mounted: function() {
        console.log(this.user)

        if (this.user == null) {
            this.$router.push('/login')
        }

        setInterval(this.getElection, 100)
        setInterval(this.getElectionFinish, 100)
        // setInterval(this.getElectionCreate, 100)
    }
};
</script>

<style lang="scss" scoped>
@import '../../node_modules/bootstrap/scss/bootstrap';
@import '../../node_modules/bootstrap-vue/src/index.scss';
@import "../../node_modules/@syncfusion/ej2-base/styles/material.css";
@import "../../node_modules/@syncfusion/ej2-layouts/styles/material.css";
@import '../../node_modules/bootstrap/scss/bootstrap';
@import '../../node_modules/bootstrap-vue/src/index.scss';
</style>

<style scoped>
#profile {
    overflow: hidden;
}

#userProfile {
    margin: 20px;
    height: 20%;
}

#leftProfile {
    width: 30%;
    height: 100%;
    float: left;
}

#rightProfile {
    width: 70%;
    height: 100%;
    float: right;
}

th, td {
padding: 5px;
text-align: left;
}

.avatar {
  vertical-align: middle;
  width: 100px;
  height: 100px;
}

#createPoll {
    margin: 20px;
}

#paticipatePoll {
    display: inline-block;
}
</style>