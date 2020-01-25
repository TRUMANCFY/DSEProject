<template>
    <div>
        <h2> Create Election </h2>
        <div id='electionBox'>
            <label> Election Name </label>
            <b-form-input
            v-model="electionName"
            placeholder="Election Name"
            trim
            ></b-form-input>
            <label> Election Description </label>
            <b-form-input
            v-model="electionDescription"
            placeholder="Election Description"
            trim
            ></b-form-input>
        </div>
        <div id='questionBox'>
            <div id='question'>
                <label for="input-live">Question List:</label>
                <b-form-textarea
                id="textarea"
                v-model="questionText"
                placeholder="Enter something..."
                rows="3"
                ></b-form-textarea>
            </div>
            <div id='submitButton'>
                <button class='btn btn-primary' @click="addChoice">Add Choice</button>
                <button class='btn btn-primary' @click="submitQuestion">Submit Question</button>
                <button class='btn btn-primary' @click="submitElection">Submit Election</button>
            </div>
        </div>
        <div id='question'>
            <label for="input-live"> New Question:</label>
            <b-form-input
            id="input-live"
            v-model="question"
            placeholder="Question"
            trim
            ></b-form-input>
        </div>
        <div id='selection' v-for="(choice, index) in choices" v-bind:key='index'>
            <label for="input-live">Choice {{ index }} :</label>
            <b-form-input
            v-model="choices[index]"
            placeholder="Choice"
            trim
            ></b-form-input>
        </div>
    </div>
</template>

<script>
import { mapState, mapActions } from 'vuex'
import config from 'config'

export default {
    data () {
        return {
            question: '',
            choices: [''],
            questionText: '',
            questionList: [],
            electionName: '',
            electionDescription: '',
        }
    },
    computed: {
        ...mapState('account', ['status', 'user'])
    },
    methods: {
        ...mapActions('account', ['register']),
        addChoice: function() {
            var self = this;
            self.choices.push('')
        },
        submitQuestion: function() {
            var self = this;
            var q = {
                question: self.question,
                choices: self.choices,
            }

            self.questionList.push(q);

            console.log(self.choices)

            var qText = 'Question' + self.questionList.length + ' :' + self.question + '\n' + 'Choices: ' + self.choices.join(' ') + '\n'

            self.questionText += qText

            self.question = ''

            self.choices = ['']
        },
        submitElection: async function() {
            var self = this;

            var electionOptions = {
                name: self.electionName,
                description: self.electionDescription,
                questions: self.questionList,
                creator: this.user.id,
                publickey: null
            }

            var payload = {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                mode: 'cors',
                body: JSON.stringify(electionOptions),
            };

            var a = await fetch(`${config.apiUrl}/createElection`, payload)
            .then(res => {
                    if (res.ok) {
                        return '';
                    }
            })

            var b = await fetch('/createElection', payload)
            .then(res => {
                if (res.ok) {
                    return ''
                }
            })

            confirm('The election has been created!')
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
#submitButton {
    margin-top: 20px;
}

</style>