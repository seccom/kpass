import { connect } from 'react-redux'
import { bindActionCreators } from 'redux'
import { reduxForm } from 'redux-form'

import { cookie } from 'utils/cookie'
import { signInUserAction } from 'modules'
import { SignIn as SignInView } from './sign-in.view'
import { signInValidate } from './sign-in.validate'

const mapStateToProps = (state) => ({})

const mapDispatchToProps = (dispatch) => ({
  actions: bindActionCreators({
    signInUser: signInUserAction
  }, dispatch)
})

const createContainer = (component) => {
  const connectedComponent = connect(
    mapStateToProps,
    mapDispatchToProps
  )(component)

  return reduxForm({
    form: 'signInForm',
    validate: signInValidate,
    initialValues: {
      username: cookie('kp_username')
    }
  })(connectedComponent)
}

export const SignIn = createContainer(SignInView)
