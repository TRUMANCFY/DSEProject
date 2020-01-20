<template>
    <div>
        <p> Name: {{ electionName }} </p>
        <p> Description: {{ electionDescription }} </p>
        <div v-for='(option, index) in options' v-bind:key='index' >
            <b-form-group>
                <label> {{ qText[index] }} </label>
                <b-form-radio-group
                    :options="option"
                    v-model="selection[index]"
                ></b-form-radio-group>
            </b-form-group>
        </div>
        <button class='btn btn-primary' @click="submitVote">Submit</button>
        <button class='btn btn-primary' @click="backToProfile">Back</button>
    </div>
</template>

<script>
import { mapState, mapActions } from 'vuex'

export default {
    data () {
        return {
            options: [],
            selection: [],
            electionName: '',
            electionDescription: '',
            qText: [],
        }
    },
    props: {
        questions: Object,
    },
    computed: {
        ...mapState('account', ['status', 'user'])
    },
    methods: {
        ...mapActions('account', ['register']),
        generateVote: function() {
            var self = this;

            var res = [];

            self.options.map((op, ind) => {
                let tmp = [self.selection[ind]];
                res.push(tmp);
            })

            console.log(res);
            return res;
        },
        submitVote: function() {
            var self = this;
            let voteRes = self.generateVote();
            
            var payload = {
                'answers': voteRes,
            }

            var a = fetch('/vote', {method: "POST", body: JSON.stringify(payload), mode: 'cors'})
            .then(res => {
                if (res.ok) {
                var temp = res.json()
                return temp
                }
            });

            confirm('The vote has been submitted!')

            this.$router.push('/users')
        },
        backToProfile: function() {
            this.$router.push('/users')
        }
    },
    mounted: function() {
        var self = this;
        console.log(this.questions)

        if (this.questions == null) {
            alert('No input')
            return
        }

        self.electionName = self.questions['name']
        self.electionDescription = self.questions['description']
        
        self.questions.questions.map(q => {
            self.qText.push(q['question'])
            self.selection.push(0)

            var opt = []
            q.choices.map((choice, ind) => {
                opt.push({
                    value: ind,
                    text: choice,
                })
            })

            self.options.push(opt);
        })
        
        
        // console.log(this.user)

        // if (this.user == null) {
        //     this.$router.push('/login')
        // }
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

</style>