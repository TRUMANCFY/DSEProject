<template>
    <div>
        <div id='profile'>
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

        <div id='poll'>
            <button class="btn btn-primary" @click="goVote">Go to vote</button>
            <button class="btn btn-primary" @click='createElection' style='margin-left: 20px;'> Create an election </button>
        </div>

        <div class='form-group' style='margin-top: 10px;'>
            <b-form-select v-model="selected" :options="elections" :select-size="6" style='width: 30%;'></b-form-select>
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

            if (newElections.length != self.elections.length) {
                self.elections = newElections;
                
                self.elections.forEach(element => {
                    element['text'] = element['name'];
                    element['value'] = element['name'];
                });
            }

            // self.elections.forEach(element => {
            //     element['text'] = element['name'];
            // });
        },
        createElection: function() {
            this.$router.push('/create')
        }
    },
    mounted: function() {
        console.log(this.user)

        if (this.user == null) {
            this.$router.push('/login')
        }

        setInterval(this.getElection, 100)
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