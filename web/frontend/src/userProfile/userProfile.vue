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
            <div id='createPoll'>
                <button class="btn btn-primary" @click="goVote">Go to vote</button>
            </div>

            <div id='paticipatePoll'>
                <button class="btn btn-primary">Register</button>
            </div>
        </div>
    </div>
</template>

<script>
import { mapState, mapActions } from 'vuex'

// import image
import registeredImage from '../assets/registeredUser.png'

export default {
    data () {
        return {
            registeredImage: registeredImage,
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
            this.$router.push('/vote')
        }
    },
    mounted: function() {
        console.log(this.user)

        if (this.user == null) {
            this.$router.push('/login')
        }
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