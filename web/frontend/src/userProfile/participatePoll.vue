<template>
    <div>
        <div v-for='(option, index) in options' v-bind:key='index' >
            <b-form-group label="Radios using options">
                <b-form-radio-group
                    :options="option"
                    v-model="selection[index]"
                ></b-form-radio-group>
            </b-form-group>
        </div>
        <button class='btn btn-primary' @click="submitVote">Submit</button>
    </div>
</template>

<script>
import { mapState, mapActions } from 'vuex'

export default {
    data () {
        return {
            options: [[
                    { text: 'A', value: 0 },
                    { text: 'B', value: 1 },
                    { text: 'C', value: 2 },
                    { text: 'D', value: 3 },
                ],[
                    { text: 'A+', value: 0 },
                    { text: 'B+', value: 1 },
                    { text: 'C+', value: 2 },
                    { text: 'D+', value: 3 },
                ],[
                    { text: 'A-', value: 0 },
                    { text: 'B-', value: 1 },
                    { text: 'C-', value: 2 },
                    { text: 'D-', value: 3 },
                ],],
            selection: [0, 0, 0],
        }
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
                let tmp = Array(op.length).fill(0);
                tmp[self.selection[ind]] = 1;
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
        }
    },
    mounted: function() {
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