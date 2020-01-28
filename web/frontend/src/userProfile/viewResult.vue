<template>
    <div>
        <h2> View Result </h2>
        <h4> Election Name: {{election}} </h4>
        <label> Questions </label>
        <b-form-select v-model="selected" :options="questions"></b-form-select>
        <b-button style='margin-top:20px' @click="updateChart"> Display Result </b-button>
        <Pie :answers="getAnswers" :data="getData" :key="index"/>
    </div>
</template>

<script>
import { mapState, mapActions } from 'vuex'
import config from 'config'
import Pie from './PieChart' 

export default {
    props: {
        questionsIn: Object,
        result: Object,
    },
    components: {
        Pie,
    },
    data () {
        return {
            questions: [],
            election: "",
            QAmap: {},
            selected: "",
            res: [[1,1],[2,1]],
            answerIn: [],
            dataIn: [],
            index: 0,
        }
    },
    computed: {
        ...mapState('account', ['status', 'user']),
        getData: function() {
            return this.dataIn
        },
        getAnswers: function() {
            return this.answerIn
        }
    },
    methods: {
        ...mapActions('account', ['register']),

        updateChart: function() {
        var self = this
        // get index
        var ind = 0;

        console.log(self.selected)

        console.log(self.questions)

        self.questions.map((d, i) => {
            if (self.selected == d) {
                ind = i;
            }
        })

        console.log(ind)

        self.res = self.result.res;

        self.dataIn = self.res[ind]

        self.answerIn = self.QAmap[self.selected]

        self.index += 1

        console.log(self.answerIn)

        console.log(self.dataIn)

    }
    },
    mounted: function() {
    // Build the chart

    var self = this;
    self.questions = []
    self.QAmap = {}

    self.election = self.questionsIn.name;
    
    
    self.questionsIn.questions.map(e => {
        self.questions.push(e.question)
        self.QAmap[e.question] = e.choices
    })


    self.selected = self.questions[0];

    console.log(self.result)

    self.updateChart()
    },
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